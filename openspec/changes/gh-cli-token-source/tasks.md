## 1. gh CLI auth provider

- [ ] 1.1 New `internal/provider/ghcli_auth.go`: `GhCliAuthProvider` with `NewGhCliAuthProvider()`, compile-time `var _ AuthProvider = (*GhCliAuthProvider)(nil)`; `Token` runs `gh auth token` via `exec.CommandContext` and returns trimmed stdout, no caching
- [ ] 1.2 Add `ErrGhNotFound` and `ErrGhUnauthenticated` to `provider.go`; `Token` maps `exec.ErrNotFound` from `exec.LookPath("gh")` to the former and a non-zero exit to the latter, each with its own remedy in the message
- [ ] 1.3 `Login` verifies gh availability and returns a clear error when unavailable; `Logout` is a no-op on gh's own credentials (tokenpile never owns them) — the sentinel removal is handled by the auth command
- [ ] 1.4 Add `GhLogin(ctx) (string, error)` reading the authenticated handle from `gh api user --jq .login` for status and login output
- [ ] 1.5 Unit tests with a fake `gh` script prepended to `PATH` in `t.TempDir()`: token returned and trimmed, two calls execute gh twice, missing binary yields `ErrGhNotFound`, non-zero exit yields `ErrGhUnauthenticated`

## 2. Token source persistence

- [ ] 2.1 Add `TokenSource` type with `TokenSourceOAuth`/`TokenSourceGhCli` and the `gh-cli:` sentinel constant to `internal/provider/`
- [ ] 2.2 `StoredTokenSource(credPath) (TokenSource, error)`: reads the keychain slot, falls back to the encrypted file, returns OAuth when the value is a real token and gh-cli when it is the sentinel
- [ ] 2.3 Export a way for the auth command to persist the sentinel through the existing `storeToken` path (keychain, encrypted-file fallback on failure)
- [ ] 2.4 `GitHubAuthProvider.Token`: return `ErrUnauthenticated` when the stored value is the sentinel, so it can never be sent as a bearer token
- [ ] 2.5 Tests over a temp `TOKENPILE_CONFIG_DIR`: sentinel round-trips, OAuth token reads back as `TokenSourceOAuth`, empty slot reads back as OAuth, sentinel makes `GitHubAuthProvider.Token` return `ErrUnauthenticated`

## 3. Composition root and auth command

- [ ] 3.1 `main.go`: resolve `StoredTokenSource` at startup and construct the matching `AuthProvider` for `GitHubIssueProvider`; pass both implementations to the auth command so `login` can switch in either direction
- [ ] 3.2 `cmd_auth.go` `login`: add `--use-gh-cli` and `--no-gh-cli` (mutually exclusive); when neither is passed and gh is installed and authenticated, prompt with the detected login defaulting to yes; gh absent or unauthenticated means no prompt and the OAuth flow
- [ ] 3.3 `login` gh path: validate the borrowed token by fetching the authenticated user before persisting; on failure exit non-zero naming the missing scope and `gh auth refresh -s repo`, leaving the stored source unchanged
- [ ] 3.4 `login` OAuth path: unchanged flow, overwriting any sentinel with the real token
- [ ] 3.5 `auth status`: print the active token source — OAuth App, or gh CLI with its login; when the source is gh and gh is missing or unauthenticated, report it unavailable with the remedy and exit non-zero
- [ ] 3.6 `auth logout`: clears the sentinel through the existing `Logout` path, leaving gh's own credentials untouched
- [ ] 3.7 Improve the org-restriction error path: when an OAuth-sourced call fails with 403, suggest `tokenpile auth login --use-gh-cli`

## 4. Integration tests

- [ ] 4.1 Auth integration tests with temp config dir and a fake `gh` on `PATH`: `--use-gh-cli` persists the sentinel without a browser, `--no-gh-cli` skips detection, gh absent runs no prompt
- [ ] 4.2 Selection survives across invocations: after `--use-gh-cli`, a fresh command run resolves the gh source
- [ ] 4.3 `auth status` output for both sources, and non-zero exit when the gh source is broken
- [ ] 4.4 `auth logout` clears the sentinel and returns status to unauthenticated
- [ ] 4.5 Assert no silent fallback: with the gh source active and gh broken, the command fails rather than running the OAuth flow

## 5. Docs and checks

- [ ] 5.1 README: auth section covering org-restricted repositories, choosing the token source, and the gh CLI prerequisites
- [ ] 5.2 CLAUDE.md: project map entry for `internal/provider/ghcli_auth.go` and a line in the key design decisions summary
- [ ] 5.3 Run `make check` and commit
