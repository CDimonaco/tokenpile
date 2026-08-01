## ADDED Requirements

### Requirement: Entries record their capture branch

A usage entry SHALL record the git branch that was checked out when it was captured. The capture path already reads it — Claude Code reports `gitBranch` on every transcript record and the opencode plugin resolves it — and SHALL carry it through to the entry instead of using it only to infer an issue.

The branch SHALL NOT be re-derived at listing time from the working directory: that reports today's branch, not the one the tokens were spent on, and would be wrong for exactly the old sessions worth reconciling.

Entries captured before this requirement SHALL keep an empty branch and offer no suggestion, rather than being backfilled with a value that cannot be known after the fact.

#### Scenario: Branch is carried from capture to storage
- **WHEN** a turn is captured on branch `feat/42-thing`
- **THEN** the stored entry records that branch

#### Scenario: Unknown branch is empty, not invented
- **WHEN** a turn is captured with no branch available
- **THEN** the stored entry records an empty branch
- **THEN** its group offers no suggestion

## MODIFIED Requirements

### Requirement: Reconciliation of unattributed usage

The system SHALL provide a way to list unattributed usage grouped by session, showing its repository, the branch it was captured on, time window, entry count, tokens and cost, and to assign a whole group to an issue in one action. Assignment SHALL be reversible, and SHALL move the session along with its entries so wall-clock time follows the tokens.

Grouping is required rather than convenient: a session is many turns, and per-entry assignment would make reconciliation unusable. The session is also the unit a person reasons about when saying "that stretch of work was issue 42".

The group key SHALL be the session alone. Assignment moves a whole session, so a group narrower than a session could not be assigned on its own: assigning one such group would silently move another's entries too. Where a session's entries span more than one branch, the group SHALL report no branch and offer no suggestion, rather than reporting one of them and suggesting an issue for work only partly done on it.

Where the stored branch implies an issue, the system SHALL offer it as a suggestion. The suggestion SHALL be derived on read rather than stored, so improving the rule does not require a migration, and SHALL NOT be applied automatically: it is shown, editable, and confirmed.

#### Scenario: Groups are listed by session
- **WHEN** unattributed entries exist across two sessions
- **THEN** they are presented as two groups, each with its repository, branch, entry count, tokens and cost

#### Scenario: Groups are offered with suggestions
- **WHEN** unattributed entries exist from branch `feat/42-foo`
- **THEN** that group is presented with issue 42 suggested

#### Scenario: A suggestion is never applied on its own
- **WHEN** a group carries a suggested issue
- **THEN** no attribution happens until the assignment is confirmed

#### Scenario: A session spanning two branches stays one group
- **WHEN** a session's unattributed entries were captured on two different branches
- **THEN** they are presented as a single group
- **THEN** that group reports no branch and offers no suggestion

#### Scenario: Bulk assignment
- **WHEN** a group of unattributed entries is assigned to issue 42
- **THEN** every entry in the group is attributed to issue 42

#### Scenario: Assignment moves the session too
- **WHEN** a session's usage is assigned to issue 42
- **THEN** the session itself is attributed to issue 42, so its wall-clock time is counted with its tokens

#### Scenario: Assignment can be undone
- **WHEN** a group assigned to issue 42 is unassigned
- **THEN** those entries return to unattributed with their token counts unchanged
