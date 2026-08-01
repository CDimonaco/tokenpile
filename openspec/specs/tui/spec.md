## Purpose

Define the terminal UI's views: the issue list, the issue detail view, and the budget indicators shown in them.

## Requirements

### Requirement: Issue list view

The issue list SHALL show a budget progress indicator for issues that have a budget set. The indicator SHALL display `$spent / $budget` with color coding:
- Green when spent < 80% of budget
- Yellow when spent is 80–99% of budget
- Red when spent >= 100% of budget

Issues without a budget SHALL show no indicator in that column.

#### Scenario: Issue list displays tracked issues
- **WHEN** the TUI is opened
- **THEN** tracked issues are listed with token totals, cost, and time
- **THEN** issues with a budget show a `$spent / $budget` indicator in the appropriate color

#### Scenario: Empty state
- **WHEN** no usage has been logged
- **THEN** an empty state message is shown

#### Scenario: Navigate to detail view
- **WHEN** the user selects an issue and presses Enter
- **THEN** the detail view for that issue is shown

#### Scenario: Budget indicator green when under 80%
- **WHEN** an issue has a $5.00 budget and $2.00 spent
- **THEN** the indicator `$2.00 / $5.00` is shown in green

#### Scenario: Budget indicator yellow when 80–99%
- **WHEN** an issue has a $5.00 budget and $4.20 spent
- **THEN** the indicator `$4.20 / $5.00` is shown in yellow

#### Scenario: Budget indicator red when over budget
- **WHEN** an issue has a $5.00 budget and $5.30 spent
- **THEN** the indicator `$5.30 / $5.00` is shown in red

### Requirement: Issue detail view

The Summary tab SHALL show a budget status block when a budget is set for the issue. It SHALL display: budget amount, amount spent, percentage, and overage if over budget.

#### Scenario: Detail view shows budget when set
- **WHEN** a budget is set for the issue
- **THEN** the Summary tab shows `Budget: $3.20 / $5.00 (64%)`

#### Scenario: Detail view shows overage when over budget
- **WHEN** spent exceeds the budget
- **THEN** the Summary tab shows the overage in red: `Budget: $5.30 / $5.00 (106%) — over by $0.30`

#### Scenario: Detail view shows no budget block when not set
- **WHEN** no budget is set for the issue
- **THEN** no budget line is shown in the Summary tab

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
