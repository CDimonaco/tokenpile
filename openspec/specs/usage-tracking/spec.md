## Purpose

Define the `tokenpile log` command that records token usage: its required and optional flags, and the entry it creates.

## Requirements

### Requirement: log command

The system SHALL provide a `tokenpile log` CLI command that creates a usage entry. The following flags SHALL be required: `--issue`, `--agent`, `--model`. At least one token flag SHALL be required from: `--input`, `--cache-write`, `--cache-read`, `--output`, `--reasoning`, each a non-negative integer defaulting to zero. The `--repo` flag is optional and inferred per the issue-provider spec if absent. The following flags SHALL be optional: `--note` (string, max 200 chars), `--tag` (string, repeatable).

Entries created by this command are `estimated`: a model cannot observe its own cache tiers, so `log` is not expected to populate them.

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
- **WHEN** `tokenpile log --issue 42 --agent claude-code --model claude-sonnet-4-6` is called with no token flag
- **THEN** the CLI exits non-zero explaining that at least one token count is required

#### Scenario: Log with note and tags
- **WHEN** `tokenpile log ... --note "fixed unicode handling" --tag refactor --tag bug`
- **THEN** the session note is updated to "fixed unicode handling"
- **THEN** "refactor" and "bug" are added to the session tags

#### Scenario: Log without note or tags is valid
- **WHEN** `tokenpile log` is called without `--note` or `--tag`
- **THEN** the entry is persisted with an empty note and no tags

### Requirement: Token usage recorded by billing tier

A usage entry SHALL record token counts split by the tier the provider bills them under, as a `Usage` value with `InputFresh`, `CacheWrite`, `CacheRead`, `Output` and `Reasoning`.

`Reasoning` SHALL be a subset of `Output` and SHALL NOT be added to it for any purpose, including cost computation and report totals. Providers that report reasoning tokens report them inside the completion count.

Entries SHALL NOT carry an aggregate `tokens_in` or `tokens_out` field. Totals shown in reports are computed from the tiers.

#### Scenario: Tiers persisted independently
- **WHEN** an entry is logged with fresh input, cache write, cache read and output counts
- **THEN** each tier is persisted separately and read back unchanged

#### Scenario: Reasoning does not inflate output
- **WHEN** one entry has 100 output tokens with 40 reasoning tokens
- **WHEN** another identical entry has 100 output tokens with 0 reasoning tokens
- **THEN** both entries produce the same cost
- **THEN** both report the same output token total

#### Scenario: Absent tiers are zero, not unknown
- **WHEN** an entry is logged with only fresh input and output
- **THEN** cache write and cache read are recorded as zero
- **THEN** the entry's cost includes no cache component

### Requirement: Entry provenance

Every usage entry SHALL record a `source` of either `measured` or `estimated`. `measured` means the counts came from an agent's own transcript as reported by the provider; `estimated` means they were declared by a model.

The value SHALL be derived from the command that writes the entry and SHALL NOT be settable by a flag. Entries written by `tokenpile log` SHALL be `estimated`.

#### Scenario: log writes estimated entries
- **WHEN** `tokenpile log` persists an entry
- **THEN** the entry's source is `estimated`

#### Scenario: Source cannot be forced
- **WHEN** `tokenpile log --source measured` is attempted
- **THEN** the CLI rejects the unknown flag
