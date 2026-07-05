## MODIFIED Requirements

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
