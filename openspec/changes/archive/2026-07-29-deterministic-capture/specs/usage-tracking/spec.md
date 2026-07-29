## MODIFIED Requirements

### Requirement: log command

The system SHALL provide a `tokenpile log` CLI command that creates a usage entry manually. The following flags SHALL be required: `--agent`, `--model`. At least one token flag SHALL be required from: `--input`, `--cache-write`, `--cache-read`, `--output`, `--reasoning`, each a non-negative integer defaulting to zero. The `--issue` flag is optional: when absent the entry is recorded unattributed, following the attribution resolution order. The `--repo` flag is optional and inferred per the issue-provider spec if absent. The following flags SHALL be optional: `--note` (string, max 200 chars), `--tag` (string, repeatable).

Entries created by this command are `estimated`: a model cannot observe its own cache tiers, so `log` is not expected to populate them. Measured entries come from the capture path, not from this command.

#### Scenario: Successful log with explicit repo
- **WHEN** `tokenpile log --issue 42 --agent claude-code --model claude-sonnet-4-6 --input 1234 --output 890 --repo owner/repo`
- **THEN** a usage entry is persisted
- **THEN** the CLI exits with code 0 and no output (machine-friendly)

#### Scenario: Successful log with inferred repo
- **WHEN** `tokenpile log --issue 42 --agent opencode --model gpt-4o --input 500 --output 200`
- **WHEN** the git remote resolves to `owner/repo`
- **THEN** a usage entry is persisted with `Repo = "owner/repo"`

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

#### Scenario: No token flag at all is rejected
- **WHEN** `tokenpile log --agent claude-code --model claude-sonnet-4-6` is called with no token flag
- **THEN** the CLI exits non-zero explaining that at least one token count is required

#### Scenario: Log without an issue
- **WHEN** `tokenpile log --agent claude-code --model claude-sonnet-4-6 --input 100 --output 50` is called with no `--issue` and no binding or inferable branch
- **THEN** the entry is persisted unattributed
- **THEN** the CLI exits with code 0

#### Scenario: Log with note and tags
- **WHEN** `tokenpile log ... --note "fixed unicode handling" --tag refactor --tag bug`
- **THEN** the session note is updated to "fixed unicode handling"
- **THEN** "refactor" and "bug" are added to the session tags

#### Scenario: Log without note or tags is valid
- **WHEN** `tokenpile log` is called without `--note` or `--tag`
- **THEN** the entry is persisted with an empty note and no tags
