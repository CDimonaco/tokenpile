## Purpose

Define which GitHub issue a captured turn belongs to: that usage can be recorded with no issue at all, the order in which an issue is resolved, binding one explicitly, and reconciling what was left unattributed.

## Requirements

### Requirement: Usage may be recorded without an issue

A usage entry SHALL be recordable with no issue. Attribution SHALL NOT be a precondition for capture: when the issue cannot be determined, the entry is recorded unattributed rather than discarded.

An unattributed entry is a measurement awaiting an annotation. A discarded one is a measurement lost.

#### Scenario: Unknown issue still records usage
- **WHEN** a turn is captured and no issue can be resolved
- **THEN** the entry is recorded with no issue
- **THEN** its token counts are complete

#### Scenario: Unattributed usage is visible
- **WHEN** unattributed entries exist
- **THEN** the system reports their existence and total rather than omitting them from every view

### Requirement: Attribution resolution order

At capture time the system SHALL resolve the issue in this order, stopping at the first that yields one:

1. an explicit binding recorded for the current session or working directory
2. inference from the current git branch name
3. none, recording the entry unattributed

Branch inference SHALL be offline and pattern based. It SHALL NOT call the GitHub API: capture must not depend on network or credentials.

#### Scenario: Explicit binding wins
- **WHEN** a binding names issue 42 and the branch name contains 7
- **THEN** the captured entry is attributed to issue 42

#### Scenario: Branch name inferred
- **WHEN** no binding exists and the branch is `feat/42-borrow-gh-credential`
- **THEN** the captured entry is attributed to issue 42

#### Scenario: Unrecognisable branch yields no attribution
- **WHEN** no binding exists and the branch is `main`
- **THEN** the entry is recorded unattributed
- **THEN** no network call is made

### Requirement: Binding an issue to a session

The system SHALL provide `tokenpile bind --issue N`, recording that the current session or working directory belongs to that issue, so subsequent captures are attributed to it. Binding again SHALL replace the previous value.

#### Scenario: Bind attributes subsequent turns
- **WHEN** `tokenpile bind --issue 42` is called and a turn is then captured
- **THEN** that entry is attributed to issue 42

#### Scenario: Rebinding switches issues
- **WHEN** a session bound to issue 42 is rebound to issue 7
- **THEN** entries captured afterwards are attributed to issue 7
- **THEN** entries captured before remain attributed to issue 42

### Requirement: Reconciliation of unattributed usage

The system SHALL provide a way to list unattributed usage grouped by session, showing its repository, entry count, tokens and cost, and to assign a whole group to an issue in one action. Assignment SHALL be reversible, and SHALL move the session along with its entries so wall-clock time follows the tokens.

Grouping is required rather than convenient: a session is many turns, and per-entry assignment would make reconciliation unusable. The session is also the unit a person reasons about when saying "that stretch of work was issue 42".

Suggesting an issue from the branch name is deliberately out of scope here: entries do not record the branch they were captured on, so a suggestion cannot be derived at listing time without first storing it.

#### Scenario: Groups are listed by session
- **WHEN** unattributed entries exist across two sessions
- **THEN** they are presented as two groups, each with its repository, entry count, tokens and cost

#### Scenario: Bulk assignment
- **WHEN** a group of unattributed entries is assigned to issue 42
- **THEN** every entry in the group is attributed to issue 42

#### Scenario: Assignment moves the session too
- **WHEN** a session's usage is assigned to issue 42
- **THEN** the session itself is attributed to issue 42, so its wall-clock time is counted with its tokens

#### Scenario: Assignment can be undone
- **WHEN** a group assigned to issue 42 is unassigned
- **THEN** those entries return to unattributed with their token counts unchanged
