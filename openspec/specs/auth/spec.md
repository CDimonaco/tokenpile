## Purpose

Define how tokenpile obtains a GitHub credential and where that credential is kept: the OAuth App flow, the gh CLI as an alternative token source for organizations that never approve the OAuth App, and the signing identity used for exports.

## Requirements

### Requirement: AuthProvider interface

The system SHALL define an `AuthProvider` interface in `internal/provider/` that abstracts OAuth authentication. Any provider implementation MUST satisfy this interface.

The interface SHALL expose:
- `Login(ctx context.Context) error` — starts the OAuth flow
- `Token(ctx context.Context) (string, error)` — returns a valid access token, refreshing if needed
- `Logout(ctx context.Context) error` — revokes and removes stored credentials

#### Scenario: Interface satisfied by GitHub implementation
- **WHEN** a `GitHubAuthProvider` is constructed
- **THEN** it satisfies the `AuthProvider` interface at compile time

### Requirement: Local OAuth callback flow

The system SHALL implement OAuth 2.0 authorization code flow with PKCE (S256) using a local HTTP callback server. When `tokenpile auth login --provider github` is invoked:
1. A local HTTP server SHALL start on an ephemeral loopback port (`127.0.0.1:0`)
2. The browser SHALL be opened to the provider's authorization URL with `redirect_uri=http://127.0.0.1:<port>/callback`, a random `state`, and an S256 `code_challenge`
3. Upon redirect, the server SHALL validate the `state`, extract the authorization code, and exchange it for an access token passing the PKCE `code_verifier`
4. Callback requests whose `state` does not match SHALL receive HTTP 400 and SHALL NOT abort the login flow in progress
5. The local server SHALL shut down after receiving the callback or after a 2-minute timeout

#### Scenario: Successful login
- **WHEN** the user runs `tokenpile auth login --provider github`
- **THEN** a browser window opens to GitHub OAuth authorization
- **THEN** after the user approves, the CLI prints "Logged in successfully"
- **THEN** the access token is stored in the OS keychain

#### Scenario: Callback timeout
- **WHEN** the user does not complete authorization within 2 minutes
- **THEN** the local server shuts down
- **THEN** the CLI exits with an error: "login timed out, please try again"

#### Scenario: Stray callback request does not abort login
- **WHEN** a request with a wrong or missing `state` reaches the callback endpoint during login
- **THEN** the request receives HTTP 400
- **THEN** the login keeps waiting for the legitimate callback until the timeout

#### Scenario: Unknown provider
- **WHEN** the user runs `tokenpile auth login --provider unknown`
- **THEN** the CLI exits with an error listing supported providers

### Requirement: Token storage in OS keychain

The system SHALL store OAuth access tokens in the OS keychain using `zalando/go-keyring`. On macOS this uses Keychain; on Linux this uses the Secret Service (libsecret). On Linux systems without a Secret Service, the system SHALL fall back to an AES-256-GCM encrypted file at `~/.config/tokenpile/credentials` with permissions 0600, whose key is derived from per-machine, per-user material (machine-id and uid).

#### Scenario: Token persists across invocations
- **WHEN** the user logs in
- **THEN** subsequent CLI invocations retrieve the token from the keychain without prompting

#### Scenario: Headless Linux fallback
- **WHEN** Secret Service is unavailable
- **THEN** the token is stored in the encrypted credentials file
- **THEN** the CLI prints a warning: "Secret Service unavailable, using encrypted file fallback"

### Requirement: Ed25519 signing identity generated on first run

The system SHALL generate an Ed25519 keypair on first run if `~/.config/tokenpile/identity.key` does not exist. The private key SHALL be written with permissions 0600 and the public key with 0644. The keypair is used exclusively for signing exports.

#### Scenario: First run generates keypair
- **WHEN** `~/.config/tokenpile/identity.key` does not exist
- **WHEN** any tokenpile command is invoked
- **THEN** a new Ed25519 keypair is generated and written to the config directory
- **THEN** the CLI prints "Generated signing identity at ~/.config/tokenpile/"

#### Scenario: Subsequent runs reuse keypair
- **WHEN** `~/.config/tokenpile/identity.key` already exists
- **WHEN** any tokenpile command is invoked
- **THEN** no new keypair is generated

### Requirement: Auth status command

The system SHALL provide `tokenpile auth status` which prints the current authentication state for all configured providers, including which token source is active.

#### Scenario: Authenticated
- **WHEN** the user runs `tokenpile auth status`
- **WHEN** a valid token exists for GitHub
- **THEN** the CLI prints the provider name and the authenticated username

#### Scenario: Status reports the OAuth token source
- **WHEN** the active token source is OAuth
- **THEN** the status line identifies the source as the tokenpile OAuth App

#### Scenario: Status reports the gh CLI token source
- **WHEN** the active token source is the gh CLI
- **THEN** the status line identifies the source as the gh CLI and shows the gh-authenticated login

#### Scenario: Status reports a broken gh token source
- **WHEN** the active token source is the gh CLI but `gh` is missing or no longer authenticated
- **THEN** the status line reports the token source as unavailable and names the remedy
- **THEN** the command exits non-zero

#### Scenario: Not authenticated
- **WHEN** the user runs `tokenpile auth status`
- **WHEN** no token exists
- **THEN** the CLI prints "Not logged in" and suggests running `tokenpile auth login`

### Requirement: gh CLI token source

The system SHALL provide a `GhCliAuthProvider` in `internal/provider/` satisfying the `AuthProvider` interface, whose `Token` obtains a GitHub access token by executing `gh auth token` and returning its trimmed standard output. The provider SHALL NOT be used to perform GitHub API calls: issue operations continue to run through `GitHubIssueProvider` and the go-github client, which receive the borrowed token exactly as they receive an OAuth token.

The provider SHALL NOT store the borrowed token. Every `Token` call SHALL invoke `gh auth token` afresh so that token refresh and expiry remain owned by `gh`.

The system SHALL define `ErrGhNotFound` and `ErrGhUnauthenticated` as sentinel errors distinct from `ErrUnauthenticated`.

#### Scenario: Interface satisfied by gh CLI implementation
- **WHEN** a `GhCliAuthProvider` is constructed
- **THEN** it satisfies the `AuthProvider` interface at compile time

#### Scenario: Issue lookup succeeds with a borrowed token
- **WHEN** the active token source is the gh CLI
- **WHEN** a command resolves an issue in an organization repository whose owner has not approved the tokenpile OAuth App
- **THEN** the issue is retrieved through the existing go-github client
- **THEN** no OAuth authorization flow is started

#### Scenario: gh is not installed
- **WHEN** `gh` is not present on `PATH`
- **WHEN** `Token` is called
- **THEN** the call fails with `ErrGhNotFound`
- **THEN** the message tells the user to install the gh CLI or run `tokenpile auth login` to use the OAuth flow

#### Scenario: gh is installed but not authenticated
- **WHEN** `gh` is present but `gh auth token` exits non-zero
- **WHEN** `Token` is called
- **THEN** the call fails with `ErrGhUnauthenticated`
- **THEN** the message tells the user to run `gh auth login`

#### Scenario: Token is not cached
- **WHEN** `Token` is called twice
- **THEN** `gh auth token` is executed twice
- **THEN** no token is written to the keychain or the credentials file

### Requirement: Token source selection and persistence

The system SHALL record which token source is active as a per-machine choice, made once during `tokenpile auth login` and persisted in the existing credential slot. When the gh CLI is the active source, the slot SHALL hold the sentinel value `gh-cli:` written through the same storage path as an OAuth token, so that the keychain, the encrypted-file fallback on headless Linux, and `logout` all behave unchanged.

`tokenpile auth login --provider github` SHALL detect an installed and authenticated `gh` and offer to use it, defaulting to yes. `--use-gh-cli` SHALL select the gh CLI without prompting and `--no-gh-cli` SHALL force the OAuth flow without prompting. When `gh` is absent or not authenticated, the OAuth flow SHALL run as it does today with no prompt.

The system SHALL resolve the stored token source once at startup in the composition root and construct the matching `AuthProvider`. The system SHALL NOT choose a token source per invocation and SHALL NOT silently fall back from one source to the other when a call fails.

`GitHubAuthProvider.Token` SHALL treat a stored sentinel value as absence of an OAuth token and return `ErrUnauthenticated`, so the sentinel can never be sent to GitHub as a bearer token.

#### Scenario: Detection offers the gh CLI
- **WHEN** `gh` is installed and authenticated
- **WHEN** the user runs `tokenpile auth login --provider github`
- **THEN** the CLI reports the detected gh login and asks whether to use it, defaulting to yes
- **THEN** accepting persists the gh CLI as the active token source without opening a browser

#### Scenario: Declining detection runs the OAuth flow
- **WHEN** the user declines the prompt, or passes `--no-gh-cli`
- **THEN** the OAuth authorization code flow with PKCE runs as specified
- **THEN** the resulting access token is stored in the credential slot

#### Scenario: Non-interactive selection
- **WHEN** the user runs `tokenpile auth login --provider github --use-gh-cli`
- **THEN** no prompt is shown and the gh CLI is persisted as the active token source

#### Scenario: gh absent means no prompt
- **WHEN** `gh` is not installed
- **WHEN** the user runs `tokenpile auth login --provider github`
- **THEN** no prompt is shown and the OAuth flow runs

#### Scenario: Choice survives across invocations
- **WHEN** the gh CLI has been selected
- **WHEN** any subsequent tokenpile command needs a token
- **THEN** the token is obtained from `gh auth token` without re-prompting

#### Scenario: Sentinel is never used as a bearer token
- **WHEN** the credential slot holds the sentinel
- **WHEN** `GitHubAuthProvider.Token` is called
- **THEN** it returns `ErrUnauthenticated` rather than the sentinel string

#### Scenario: No silent fallback on authorization failure
- **WHEN** the active token source is OAuth and a GitHub call fails because the organization has not approved the OAuth App
- **THEN** the command reports that failure
- **THEN** the gh CLI is not consulted, and the error suggests `tokenpile auth login --use-gh-cli`

### Requirement: Borrowed token validated at login

When the gh CLI is selected as the token source, `tokenpile auth login` SHALL validate the borrowed token by fetching the authenticated user from the GitHub API before persisting the choice. If validation fails, the command SHALL exit non-zero and SHALL NOT persist the token source.

#### Scenario: Insufficient scope refused at login
- **WHEN** the gh credential lacks the `repo` scope
- **WHEN** the user runs `tokenpile auth login --provider github --use-gh-cli`
- **THEN** the command exits non-zero naming the missing scope and suggesting `gh auth refresh -s repo`
- **THEN** the active token source is left unchanged

#### Scenario: Valid credential persisted
- **WHEN** the gh credential is accepted by the GitHub API
- **THEN** the sentinel is written to the credential slot
- **THEN** the CLI prints the authenticated login
