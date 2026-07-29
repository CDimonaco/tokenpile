package provider

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/google/go-github/v68/github"
)

// TokenInfo describes a token as GitHub sees it.
//
// ScopesKnown distinguishes "this token has no scopes" from "GitHub did not
// tell us": fine-grained PATs carry no X-OAuth-Scopes header at all, and
// treating that silence as a missing scope would reject perfectly valid
// credentials, which are exactly the ones users fall back to when their org
// restricts OAuth Apps.
type TokenInfo struct {
	Login       string
	Scopes      []string
	ScopesKnown bool
}

// HasScope reports whether the token is known to carry the given scope. It
// returns true when scopes are unknown, since an unknown scope set cannot be
// shown to be insufficient.
func (t TokenInfo) HasScope(scope string) bool {
	if !t.ScopesKnown {
		return true
	}

	return slices.Contains(t.Scopes, scope)
}

// ValidateToken fetches the authenticated user, proving the token works before
// tokenpile commits to it.
func ValidateToken(ctx context.Context, token string) (TokenInfo, error) {
	return ValidateTokenWithURL(ctx, token, "")
}

// ValidateTokenWithURL is ValidateToken against an explicit API base URL.
func ValidateTokenWithURL(ctx context.Context, token, baseURL string) (TokenInfo, error) {
	client := github.NewClient(nil).WithAuthToken(token)

	if baseURL != "" {
		var err error

		client, err = client.WithEnterpriseURLs(baseURL+"/", baseURL+"/")
		if err != nil {
			return TokenInfo{}, fmt.Errorf("set base URL: %w", err)
		}
	}

	user, resp, err := client.Users.Get(ctx, "")
	if err != nil {
		return TokenInfo{}, fmt.Errorf("validate token: %w", err)
	}

	info := TokenInfo{Login: user.GetLogin()}

	if resp != nil && resp.Response != nil {
		info.Scopes, info.ScopesKnown = parseScopes(resp.Response.Header)
	}

	return info, nil
}

func parseScopes(header http.Header) ([]string, bool) {
	raw, ok := header[http.CanonicalHeaderKey("X-OAuth-Scopes")]
	if !ok {
		return nil, false
	}

	var scopes []string

	for part := range strings.SplitSeq(strings.Join(raw, ","), ",") {
		if s := strings.TrimSpace(part); s != "" {
			scopes = append(scopes, s)
		}
	}

	return scopes, true
}
