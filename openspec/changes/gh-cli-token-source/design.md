## Context

`AuthProvider` (`internal/provider/provider.go`) is a three-method interface — `Login`, `Token`, `Logout` — and `GitHubIssueProvider.client()` consumes exactly one of them: `Token(ctx) (string, error)`. Everything GitHub-specific downstream (go-github, pagination in `ListIssues`, the 404 → `ErrIssueNotFound` mapping in `GetIssue`) sits behind that single string. That is the seam this change uses.

Credentials today live in one slot: `keyring.Get("tokenpile", "github-token")`, falling back to an AES-256-GCM file at `credPath` on headless Linux (`storeToken` / `loadEncryptedToken`). `Logout` clears both. There is no general config file — `config.Paths` resolves `pricing.yaml` (an override) and the identity keypair, nothing that stores preferences.

## Goals / Non-Goals

**Goals:**
- Make tokenpile usable against repositories in organizations that will not approve its OAuth App.
- Reuse the entire existing GitHub client path; add a token source, not a second API client.
- Keep the active token source explicit and inspectable — never ambiguous about which credential answered.
- Introduce no new config file.

**Non-Goals:**
- `GITHUB_TOKEN` / `GH_TOKEN` environment token source. The seam introduced here makes it a ~5-line addition and it is the natural way to cover CI, but it is a separate decision with its own precedence questions; deliberately deferred.
- GitHub Enterprise host selection (`gh auth token --hostname`). `NewGitHubIssueProviderWithURL` exists but is not wired in `main.go`, so there is no GHE path to extend yet.
- Using `gh` for API calls (`gh issue list --json`, `gh api`).
- Automatic re-authentication of `gh` itself. If `gh` is not logged in, tokenpile reports it and points at `gh auth login`.

## Decisions

**1. `gh` is a token source, not an issue provider.**
`GhCliAuthProvider` implements `AuthProvider`; `Token` runs `gh auth token` and returns stdout trimmed. `GitHubIssueProvider` is constructed with it exactly as with the OAuth provider and does not change by one line.
Alternative considered: implementing `IssueProvider` over `gh issue list --json` / `gh api`. Rejected — it duplicates pagination and error mapping that go-github already provides, requires hand-rolling the `ErrIssueNotFound` mapping, and multiplies the surface for no capability gain, since both paths ultimately carry the same token.

**2. Selection is persisted once, not decided per call.**
Rejected: a silent runtime fallback (try OAuth, on failure try `gh`). Two reasons. First, tokenpile's value proposition is auditable signed output; a tool that silently picks *which identity* queried GitHub works against that posture and makes "works on my machine" undiagnosable. Second, the fallback conflates two genuinely different failures — "not authenticated" and "authenticated but the org returns 403" — and hides exactly the signal the user needs.
Rejected: a per-invocation `--use-gh` flag. The primary caller is an agent invoking the skill; a flag not present in the skill prompt is a flag that is never passed.
Adopted: auto-detection at `login` time, surfaced to the user, then persisted. The convenience of a fallback, spent once, in the open.

**3. The sentinel lives in the existing credential slot.**
`login` under `gh` mode stores the literal string `gh-cli:` through the existing `storeToken` path, so it lands in the keychain or the encrypted file with the same semantics. Consequences, all of them wanted: `logout` already clears it; the headless-Linux fallback is inherited; nothing new to back up; `reset` already destroys it.
Alternative considered: a new `~/.config/tokenpile/auth.yaml`. Rejected for now — a whole config file, its loader, its precedence rules and its `reset` integration to carry one enum value.
Sentinel format is `gh-cli:` with a trailing colon so the value can later grow a parameter (a hostname, say) without a migration.

**4. Resolution happens in the composition root, once at startup.**
`provider.StoredTokenSource(credPath) (TokenSource, error)` reads the slot and reports `TokenSourceOAuth` or `TokenSourceGhCli`. `main.go` reads it and constructs the matching `AuthProvider`, which then flows into `GitHubIssueProvider` as it does today.
Alternative considered: a dispatching `AuthProvider` that resolves the mode inside every `Token` call. Rejected — it is an indirection layer whose only job is to hide a decision the composition root is supposed to make, and CLAUDE.md is explicit that the dependency graph stays explicit.
`cmd_auth.go` is the exception: `login` must be able to switch modes in either direction, so the `auth` command receives both implementations and picks per invocation. Every other command receives the single resolved one.

**5. `GitHubAuthProvider.Token` must never return the sentinel.**
If the slot holds `gh-cli:` and something still calls the OAuth provider's `Token`, returning that string would send `Authorization: bearer gh-cli:` to GitHub and surface as a baffling 401. `Token` therefore treats a sentinel value as `ErrUnauthenticated`. This is defence in depth: correct wiring means it cannot happen, but the failure mode is expensive enough to close.

**6. The `gh` token is never cached.**
`Token` shells out on every call. It costs a few milliseconds per invocation and buys correctness: `gh` owns refresh and expiry, and duplicating a credential we do not control into our own store means inheriting an expiry we cannot see. It also keeps `logout` honest — there is no copy of someone else's token left behind.

**7. Validation at login, not at first use.**
`gh` may be authenticated with a credential lacking `repo` scope, which fails only later as a 403 on an issue lookup. `login --use-gh-cli` therefore fetches the authenticated user with the borrowed token before persisting the sentinel, and refuses on failure. The error names the missing scope and the `gh auth refresh -s repo` remedy.

**8. Error taxonomy.**
`ErrGhNotFound` (`gh` not on `PATH`) and `ErrGhUnauthenticated` (`gh` present, `gh auth token` non-zero) are distinct sentinels, both distinct from `ErrUnauthenticated`. They are separate user situations with separate remedies — install `gh`, run `gh auth login`, run `tokenpile auth login` — and collapsing them produces the "not authenticated: run tokenpile auth login" message in a case where that command is precisely the wrong advice.

## Risks / Trade-offs

- [tokenpile's behaviour now depends on an external binary it does not version] → confined to one process call with an explicit error taxonomy; the OAuth path stays fully functional and is still the default when `gh` is absent.
- [Shelling out on every `Token` call adds latency to `ListIssues`/`GetIssue`] → single-digit milliseconds against a network round-trip; the TUI already amortises issue lookups.
- [The sentinel occupies a slot typed as "token"] → mitigated by decision 5; the value is syntactically impossible to confuse with a real GitHub token.
- [Users may not realise which identity produced their data] → `auth status` names the active source and the `gh` login; the choice is never made behind their back.
- [`gh` stores its token in its own keychain entry, so tokenpile inherits whatever scopes `gh` holds] → validated at login (decision 7), and documented as the explicit trade of borrowing a credential.
