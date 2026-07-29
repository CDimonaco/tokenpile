## MODIFIED Requirements

### Requirement: Skill file content

The skill template SHALL NOT instruct the agent to estimate or report token counts. Token counts come from the capture path, which reads what the provider reported; a model cannot observe its own context or cache and any figure it supplies is a guess.

The skill's remaining responsibilities are to declare the issue being worked on by calling `tokenpile bind --issue N` once it is known, and to answer questions about usage data by running CLI commands. The template SHALL instruct the agent to pass `--note` and `--tag` when binding. The note SHALL be a single line (max ~100 chars) summarizing what is being worked on. Tags SHALL be chosen from a documented vocabulary: `refactor`, `debug`, `feature`, `test`, `docs`, `spike`, `review`.

Failing to bind SHALL cost attribution only: usage is still captured, unattributed, and can be assigned later.

#### Scenario: Skill file provides correct CLI invocation
- **WHEN** the skill template for any agent is rendered
- **THEN** the example invocation is `tokenpile bind` with `--issue`, `--note` and `--tag` flags
- **THEN** the template documents the recommended tag vocabulary

#### Scenario: Template does not ask for token counts
- **WHEN** the skill template for any agent is rendered
- **THEN** it contains no instruction to estimate, count or pass token figures

#### Scenario: Agent declares the issue
- **WHEN** an agent follows the installed skill instructions and learns the issue it is working on
- **THEN** it calls `tokenpile bind` with that issue number
- **THEN** the call includes a `--note` and one or more `--tag` values from the vocabulary

#### Scenario: Not binding loses attribution, not measurement
- **WHEN** an agent never calls `tokenpile bind` during a session
- **THEN** the session's usage is still captured with complete token counts
- **THEN** those entries are unattributed and available for later assignment
