package provider

import (
	"errors"
	"fmt"
	"strings"
)

// TokenSource names where tokenpile gets its GitHub token from. The choice is
// made once at login and persisted, never decided per invocation: a tool whose
// output is signed and auditable must not be ambiguous about which credential
// answered.
type TokenSource string

const (
	TokenSourceOAuth TokenSource = "oauth"
	TokenSourceGhCli TokenSource = "gh-cli"
)

// ghCliSentinel marks the credential slot as "the gh CLI holds the real token".
// The trailing colon leaves room for a parameter (a hostname, say) without a
// storage migration. No GitHub token can collide with it.
const ghCliSentinel = "gh-cli:"

func isTokenSourceSentinel(value string) bool {
	return strings.HasPrefix(value, ghCliSentinel)
}

// StoredTokenSource reports which token source this machine is configured for.
// An empty or unreadable slot means OAuth: the user is simply not logged in
// yet, and the OAuth flow is what login runs by default.
func StoredTokenSource(credPath string) TokenSource {
	value, err := loadCredential(credPath)
	if err != nil {
		return TokenSourceOAuth
	}

	if isTokenSourceSentinel(value) {
		return TokenSourceGhCli
	}

	return TokenSourceOAuth
}

// PersistGhCliTokenSource records the gh CLI as this machine's token source. It
// writes only the marker: the token itself stays with gh.
func PersistGhCliTokenSource(credPath string) error {
	if err := storeCredential(credPath, ghCliSentinel); err != nil {
		return fmt.Errorf("persist token source: %w", err)
	}

	return nil
}

// IsGhCredentialError reports whether an error comes from the gh CLI being
// missing or unusable, so callers can tell a borrowed-credential problem from
// an ordinary auth problem.
func IsGhCredentialError(err error) bool {
	return errors.Is(err, ErrGhNotFound) || errors.Is(err, ErrGhUnauthenticated)
}
