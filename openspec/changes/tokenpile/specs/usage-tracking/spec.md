## ADDED Requirements

### Requirement: Store interface

The system SHALL define a `Store` interface in `internal/store/` that abstracts all persistence. Domain logic and CLI commands MUST interact with storage only through this interface. The interface SHALL expose at minimum:

- `LogUsage(ctx context.Context, entry UsageEntry) error`
- `ListIssues(ctx context.Context, filter IssueFilter) ([]TrackedIssue, error)`
- `GetReport(ctx context.Context, repo string, issueNum int) (*Report, error)`
- `StartSession(ctx context.Context, repo string, issueNum int) (*Session, error)`
- `EndSession(ctx context.Context, sessionID string) error`
- `ListSessions(ctx context.Context, repo string, issueNum int) ([]Session, error)`

#### Scenario: SQLite adapter satisfies Store interface
- **WHEN** a `SQLiteStore` is constructed with a valid DB path
- **THEN** it satisfies the `Store` interface at compile time

### Requirement: UsageEntry domain type

The system SHALL define a `UsageEntry` type in `internal/domain/` with the following fields:
- `ID` string (UUID, generated on insert)
- `Repo` string (owner/repo)
- `IssueNum` int
- `Agent` string (required, e.g. "claude-code", "opencode")
- `Model` string (required, e.g. "claude-sonnet-4-6", "gpt-4o")
- `TokensIn` int
- `TokensOut` int
- `SessionID` string (foreign key to sessions table)
- `At` time.Time (UTC, set on insert)

#### Scenario: Entry stored with all required fields
- **WHEN** `LogUsage` is called with a fully populated `UsageEntry`
- **THEN** the entry is persisted and retrievable

### Requirement: log command

The system SHALL provide a `tokenpile log` CLI command that creates a `UsageEntry`. The following flags SHALL be required: `--issue`, `--agent`, `--model`, `--tokens-in`, `--tokens-out`. The `--repo` flag is optional and inferred per the issue-provider spec if absent.

#### Scenario: Successful log with explicit repo
- **WHEN** `tokenpile log --issue 42 --agent claude-code --model claude-sonnet-4-6 --tokens-in 1234 --tokens-out 890 --repo owner/repo`
- **THEN** a `UsageEntry` is persisted
- **THEN** the CLI exits with code 0 and no output (machine-friendly)

#### Scenario: Successful log with inferred repo
- **WHEN** `tokenpile log --issue 42 --agent opencode --model gpt-4o --tokens-in 500 --tokens-out 200`
- **WHEN** the git remote resolves to `owner/repo`
- **THEN** a `UsageEntry` is persisted with `Repo = "owner/repo"`

#### Scenario: Missing required flag
- **WHEN** `tokenpile log` is called without `--agent`
- **THEN** the CLI exits with a non-zero code and an error message naming the missing flag

#### Scenario: Agent name is required and not inferred
- **WHEN** `tokenpile log` is called without `--agent`
- **THEN** the CLI does not attempt to infer the agent name
- **THEN** it exits with an error: "flag --agent is required"

#### Scenario: Model name is required and not inferred
- **WHEN** `tokenpile log` is called without `--model`
- **THEN** the CLI exits with an error: "flag --model is required"

### Requirement: SQLite schema

The system SHALL create the following tables on first open if they do not exist:

```sql
CREATE TABLE IF NOT EXISTS usage_entries (
    id          TEXT PRIMARY KEY,
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    agent       TEXT NOT NULL,
    model       TEXT NOT NULL,
    tokens_in   INTEGER NOT NULL,
    tokens_out  INTEGER NOT NULL,
    session_id  TEXT,
    at          TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sessions (
    id          TEXT PRIMARY KEY,
    repo        TEXT NOT NULL,
    issue_num   INTEGER NOT NULL,
    started_at  TEXT NOT NULL,
    ended_at    TEXT
);
```

#### Scenario: Schema created on first run
- **WHEN** the SQLite database file does not exist
- **WHEN** any Store method is called
- **THEN** the database file is created and the schema is applied

#### Scenario: Schema is idempotent
- **WHEN** the database already exists with the correct schema
- **WHEN** the Store is constructed again
- **THEN** no error occurs and existing data is preserved
