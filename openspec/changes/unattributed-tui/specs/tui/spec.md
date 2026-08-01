## ADDED Requirements

### Requirement: Unattributed usage view

The TUI SHALL provide a view listing unattributed usage grouped by session, showing repository, branch, time window, entry count, tokens, cost and the suggested issue where one is implied.

The view SHALL allow assigning the selected group to an issue, with any suggestion pre-filled and editable before confirmation, and unassigning a group that was assigned. Both SHALL use the same store operations as the CLI, so the two surfaces cannot disagree.

#### Scenario: View lists groups
- **WHEN** unattributed usage exists and the view is opened
- **THEN** each group shows its repository, branch, time window, entries, tokens and cost

#### Scenario: Suggestion is pre-filled and editable
- **WHEN** a selected group's branch implies issue 42
- **THEN** the assignment prompt opens pre-filled with 42
- **THEN** the number can be changed before confirming

#### Scenario: Assigning updates the list
- **WHEN** a group is assigned to an issue
- **THEN** it disappears from the unattributed list
- **THEN** its usage appears under that issue

#### Scenario: Empty state
- **WHEN** no unattributed usage exists
- **THEN** the view says so rather than showing an empty table

### Requirement: Unattributed usage is discoverable

Unattributed usage counts toward no budget, so it is invisible spending until someone looks for it. The issue list SHALL indicate when unattributed usage exists and offer a key to open the view.

#### Scenario: Issue list signals unattributed usage
- **WHEN** unattributed usage exists for the repository in view
- **THEN** the issue list shows that it exists and how to open the view

#### Scenario: No signal when nothing is unattributed
- **WHEN** no unattributed usage exists
- **THEN** the issue list shows no such indicator
