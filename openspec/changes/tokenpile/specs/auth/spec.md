## ADDED Requirements

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

The system SHALL implement OAuth 2.0 authorization code flow using a local HTTP callback server. When `tokenpile auth login --provider github` is invoked:
1. A local HTTP server SHALL start on an ephemeral port
2. The browser SHALL be opened to the provider's authorization URL with `redirect_uri=http://localhost:<port>/callback`
3. Upon redirect, the server SHALL extract the authorization code and exchange it for an access token
4. The local server SHALL shut down after receiving the callback or after a 2-minute timeout

#### Scenario: Successful login
- **WHEN** the user runs `tokenpile auth login --provider github`
- **THEN** a browser window opens to GitHub OAuth authorization
- **THEN** after the user approves, the CLI prints "Logged in successfully"
- **THEN** the access token is stored in the OS keychain

#### Scenario: Callback timeout
- **WHEN** the user does not complete authorization within 2 minutes
- **THEN** the local server shuts down
- **THEN** the CLI exits with an error: "login timed out, please try again"

#### Scenario: Unknown provider
- **WHEN** the user runs `tokenpile auth login --provider unknown`
- **THEN** the CLI exits with an error listing supported providers

### Requirement: Token storage in OS keychain

The system SHALL store OAuth access tokens in the OS keychain using `zalando/go-keyring`. On macOS this uses Keychain; on Linux this uses the Secret Service (libsecret). On Linux systems without a Secret Service, the system SHALL fall back to an AES-256-GCM encrypted file at `~/.config/tokenpile/credentials` with permissions 0600.

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

The system SHALL provide `tokenpile auth status` which prints the current authentication state for all configured providers.

#### Scenario: Authenticated
- **WHEN** the user runs `tokenpile auth status`
- **WHEN** a valid token exists for GitHub
- **THEN** the CLI prints the provider name and the authenticated username

#### Scenario: Not authenticated
- **WHEN** the user runs `tokenpile auth status`
- **WHEN** no token exists
- **THEN** the CLI prints "Not logged in" and suggests running `tokenpile auth login`
