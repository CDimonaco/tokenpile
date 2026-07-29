## Why

tokenpile authenticates through a GitHub OAuth App. When the target repository belongs to an organization, OAuth App access must be approved by an org admin, and in many organizations that approval is never granted. Today this makes tokenpile unusable against those repositories: `GetIssue` and `ListIssues` fail even though the developer has perfectly valid access to the repo from their terminal.

The `gh` CLI does not bypass those restrictions, but its credential has almost always already been granted: `gh` is whitelisted in practically every organization because developers cannot work without it, and where it is not, `gh auth login --with-token` accepts a fine-grained PAT, which org OAuth restrictions do not govern at all. So tokenpile can inherit the credential the developer already fought to obtain instead of asking for a second approval that will not come.

## What Changes

- New `GhCliAuthProvider` in `internal/provider/`: an `AuthProvider` implementation whose `Token` shells out to `gh auth token`. `gh` is used strictly as a **token source**, never as an API client — `GitHubIssueProvider`, go-github, pagination and error mapping are untouched.
- The token source becomes a persisted, per-machine choice made once at `tokenpile auth login`, not a per-invocation flag and not a runtime fallback. `auth login` auto-detects an authenticated `gh` and offers it; `--use-gh-cli` / `--no-gh-cli` make the choice non-interactively.
- The choice is recorded as a sentinel value (`gh-cli:`) in the existing credential slot (keychain, or the encrypted file fallback). No new config file is introduced.
- `auth status` reports which token source is active and, for `gh`, the `gh`-authenticated login. `auth logout` clears the sentinel, returning the machine to the unauthenticated state.
- The `gh` token is never copied into the keychain: `Token` invokes `gh auth token` on every call so refreshes and expiries stay owned by `gh`.
- `auth login --use-gh-cli` validates the borrowed token against the GitHub API before persisting the choice, so a `gh` credential lacking `repo` scope fails at login rather than at first use.

## Capabilities

### Modified Capabilities

- `auth`: new requirements for the `gh` CLI token source — provider implementation, token-source selection and persistence, validation at login, status and logout behaviour. The existing OAuth flow requirement is unchanged; the "Auth status command" requirement gains token-source reporting.

## Impact

- `internal/provider/`: new `ghcli_auth.go` (`GhCliAuthProvider`, `ErrGhNotFound`, `ErrGhUnauthenticated`) and a `StoredTokenSource` helper reading the sentinel; `github_auth.go` gains sentinel-aware `Token` handling so the sentinel is never returned as a token.
- `cmd/tokenpile/main.go`: composition root resolves the stored token source at startup and wires the matching `AuthProvider`; the `auth` command receives both implementations so `login` can switch between them.
- `cmd/tokenpile/cmd_auth.go`: `--use-gh-cli` / `--no-gh-cli` flags, detection and prompt in `login`, token-source line in `status`.
- No change to `internal/provider/github_issues.go`, to the `IssueProvider` interface, or to the generated mocks.
- Tests: `GhCliAuthProvider` unit tests with a stubbed `gh` binary on `PATH`; auth command integration tests for selection, persistence, status and logout.
- Docs: README auth section (org-restricted repositories, choosing the token source), CLAUDE.md project map entry for `ghcli_auth.go`.
