## ADDED Requirements

### Requirement: Measured capture from agent transcripts

The system SHALL record token usage from the record the agent itself keeps of what the provider reported, never from a count estimated by a model. Entries created this way SHALL have `source = measured` and SHALL populate the usage tiers.

Capture SHALL be triggered by a mechanism the agent invokes regardless of model behaviour, and SHALL NOT depend on the model choosing to act.

#### Scenario: A turn is captured without model cooperation
- **WHEN** an agent completes a turn and the model issues no tokenpile command
- **THEN** the turn's usage is recorded
- **THEN** the entry's source is `measured`

#### Scenario: Captured counts come from the provider
- **WHEN** a turn's transcript reports 609 fresh input, 477,012 cache write, 43,907,365 cache read and 225,365 output tokens
- **THEN** the recorded entry carries exactly those per-tier counts

### Requirement: Claude Code capture hook

The system SHALL provide `tokenpile hook claude-code`, reading the hook payload as JSON on stdin and using `transcript_path` to read the turn's usage and model. It SHALL be installed as a `Stop` hook in the user's Claude Code settings file, not in skill frontmatter: hooks declared in skill frontmatter are scoped to the skill's lifecycle and run only while the model has that skill active, which is the non-determinism this capability exists to remove.

Claude Code reports no separate reasoning count; captured entries SHALL record reasoning as zero rather than absent, since thinking tokens are billed inside output.

#### Scenario: Stop hook records the turn
- **WHEN** the `Stop` hook runs with a payload naming a transcript containing one assistant turn
- **THEN** an entry is recorded with that turn's per-tier usage and its model

#### Scenario: Hook is installed in settings, not frontmatter
- **WHEN** the capture hook is installed for Claude Code
- **THEN** a `Stop` hook entry exists in the settings file
- **THEN** the skill's frontmatter declares no hooks

#### Scenario: Existing settings are preserved
- **WHEN** the settings file already contains unrelated hooks and settings
- **WHEN** the capture hook is installed
- **THEN** the tokenpile hook entry is added
- **THEN** every pre-existing hook and setting is unchanged

### Requirement: opencode capture plugin

The system SHALL install an opencode plugin subscribing to `session.idle`, which reads per-message usage — input, output, reasoning, cache write and cache read — and records it through tokenpile.

#### Scenario: Session idle records its messages
- **WHEN** an opencode session goes idle after producing assistant messages with usage
- **THEN** an entry is recorded per message with its per-tier counts and model

#### Scenario: Reasoning tokens are carried through
- **WHEN** an opencode message reports reasoning tokens
- **THEN** the recorded entry carries them in the reasoning tier
- **THEN** they are not added to the output tier

### Requirement: Capture is spooled before it is stored

Capture SHALL append the observed usage to an append-only spool before any database write, and a reconciler SHALL fold spooled records into the store on subsequent tokenpile invocations. A capture invocation that cannot reach the database SHALL still leave a durable record.

An agent's hook semantics let a failing hook exit non-zero and continue, so a capture path that wrote directly to the database would lose turns silently.

#### Scenario: Database unavailable does not lose the turn
- **WHEN** the capture hook runs while the database cannot be written
- **THEN** the usage is appended to the spool
- **THEN** a later tokenpile invocation records it in the store

#### Scenario: Spool is drained exactly once
- **WHEN** the reconciler processes a spool containing records already stored
- **THEN** no duplicate entries are created

#### Scenario: Unparseable input is kept, not discarded
- **WHEN** a transcript cannot be parsed
- **THEN** the raw payload is spooled for later inspection
- **THEN** the failure is reported rather than passing silently

### Requirement: Capture installation follows skill installation

`tokenpile skill install <agent>` SHALL install both the skill and that agent's capture mechanism; `tokenpile skill uninstall <agent>` and `tokenpile reset` SHALL remove both.

#### Scenario: Install sets up capture
- **WHEN** `tokenpile skill install claude-code` is called
- **THEN** the skill file is written and the capture hook is registered

#### Scenario: Uninstall removes capture
- **WHEN** `tokenpile skill uninstall claude-code` is called
- **THEN** the skill file is removed and the tokenpile hook entry is gone from the settings file
- **THEN** unrelated hooks in that file remain
