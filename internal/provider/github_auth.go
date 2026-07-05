package provider

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	keychainService = "tokenpile"
	keychainKey     = "github-token"
	oauthTimeout    = 2 * time.Minute
)

type GitHubAuthProvider struct {
	clientID     string
	clientSecret string
	credPath     string
	oauthCfg     *oauth2.Config
}

func NewGitHubAuthProvider(clientID, clientSecret, credPath string) *GitHubAuthProvider {
	return &GitHubAuthProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		credPath:     credPath,
		oauthCfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"read:user", "repo"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *GitHubAuthProvider) Login(ctx context.Context) error {
	// Ephemeral port: GitHub ignores the port on loopback redirect URLs, and a
	// runtime-chosen port prevents a local process from squatting the callback.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("start callback server: %w", err)
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		_ = listener.Close()
		return fmt.Errorf("unexpected listener address type %T", listener.Addr())
	}

	p.oauthCfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", addr.Port)

	state := randomState()
	verifier := oauth2.GenerateVerifier()
	authURL := p.oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.S256ChallengeOption(verifier))

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{ReadHeaderTimeout: 10 * time.Second}
	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}

		if got := r.URL.Query().Get("state"); got != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			errCh <- errors.New("oauth state mismatch")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("oauth callback missing code")
			return
		}

		fmt.Fprintln(w, "Login successful. You can close this tab.")
		codeCh <- code
	})

	go func() {
		if serveErr := srv.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			errCh <- serveErr
		}
	}()

	if err = openBrowser(authURL); err != nil {
		slog.Warn("could not open browser automatically", "url", authURL, "err", err)
		fmt.Printf("Open this URL in your browser:\n%s\n", authURL)
	}

	timer := time.NewTimer(oauthTimeout)
	defer timer.Stop()

	var code string

	select {
	case code = <-codeCh:
	case err = <-errCh:
		_ = srv.Shutdown(ctx)
		return fmt.Errorf("oauth callback: %w", err)
	case <-timer.C:
		_ = srv.Shutdown(ctx)
		return errors.New("login timed out, please try again")
	case <-ctx.Done():
		_ = srv.Shutdown(ctx)
		return ctx.Err()
	}

	_ = srv.Shutdown(ctx)

	token, err := p.oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return fmt.Errorf("exchange oauth code: %w", err)
	}

	if err = p.storeToken(token.AccessToken); err != nil {
		return fmt.Errorf("store token: %w", err)
	}

	return nil
}

func (p *GitHubAuthProvider) Token(_ context.Context) (string, error) {
	tok, err := keyring.Get(keychainService, keychainKey)
	if err == nil {
		return tok, nil
	}

	tok, err = p.loadEncryptedToken()
	if err != nil {
		return "", ErrUnauthenticated
	}

	return tok, nil
}

func (p *GitHubAuthProvider) Logout(_ context.Context) error {
	if err := keyring.Delete(keychainService, keychainKey); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		slog.Warn("could not delete token from keychain", "err", err)
	}

	if err := os.Remove(p.credPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials file: %w", err)
	}

	return nil
}

func (p *GitHubAuthProvider) storeToken(token string) error {
	err := keyring.Set(keychainService, keychainKey, token)
	if err == nil {
		return nil
	}

	slog.Warn("Secret Service unavailable, using encrypted file fallback", "err", err)

	return p.storeEncryptedToken(token)
}

func (p *GitHubAuthProvider) storeEncryptedToken(token string) error {
	key := machineKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(token), nil)

	if err = os.WriteFile(p.credPath, ciphertext, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

func (p *GitHubAuthProvider) loadEncryptedToken() (string, error) {
	data, err := os.ReadFile(p.credPath)
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}

	key := machineKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("credentials file corrupted")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt credentials: %w", err)
	}

	return string(plaintext), nil
}

func machineKey() []byte {
	hostname, _ := os.Hostname()
	sum := sha256.Sum256([]byte("tokenpile-v1:" + hostname))

	return sum[:]
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	return hex.EncodeToString(b)
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
