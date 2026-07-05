## MODIFIED Requirements

### Requirement: log command

The system SHALL provide a `tokenpile log` CLI command that creates a `UsageEntry`. The following flags SHALL be required: `--issue`, `--agent`, `--model`, `--tokens-in`, `--tokens-out`. The `--repo` flag is optional and inferred per the issue-provider spec if absent. The following flags SHALL be optional: `--note` (string, max 200 chars), `--tag` (string, repeatable).

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

#### Scenario: Log with note and tags
- **WHEN** `tokenpile log ... --note "fixed unicode handling" --tag refactor --tag bug`
- **THEN** the session note is updated to "fixed unicode handling"
- **THEN** "refactor" and "bug" are added to the session tags

#### Scenario: Log without note or tags is valid
- **WHEN** `tokenpile log` is called without `--note` or `--tag`
- **THEN** the log succeeds and the session annotations are unchanged
