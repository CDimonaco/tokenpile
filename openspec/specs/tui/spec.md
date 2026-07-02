
### Requirement: TUI entry point

The system SHALL launch the Bubble Tea TUI when `tokenpile` is invoked with no arguments. The TUI SHALL require the user to be authenticated; if not, it SHALL display an error message and a prompt to run `tokenpile auth login`.

#### Scenario: TUI launches on bare invocation
- **WHEN** the user runs `tokenpile` with no subcommand
- **THEN** the Bubble Tea TUI starts in the issue list view

#### Scenario: TUI shows auth error when unauthenticated
- **WHEN** the user runs `tokenpile` with no subcommand
- **WHEN** no valid OAuth token exists
- **THEN** the TUI displays "Not authenticated. Run: tokenpile auth login --provider github"

### Requirement: Issue list view

The TUI SHALL display a scrollable list of all tracked issues (issues with at least one `UsageEntry`). Each row SHALL show: issue number, repository, title (if available from provider), total cost, total tokens in+out, and total wall-clock time. The list SHALL be sortable by cost, tokens, and time. The user SHALL be able to navigate with arrow keys and press Enter to open the detail view.

#### Scenario: Issue list displays tracked issues
- **WHEN** the TUI is in issue list view
- **THEN** each tracked issue is shown as a row with number, repo, cost, tokens, and time

#### Scenario: Empty state
- **WHEN** no usage has been logged
- **THEN** the list view shows "No usage tracked yet. Run tokenpile log to get started."

#### Scenario: Navigate to detail view
- **WHEN** the user presses Enter on an issue row
- **THEN** the TUI transitions to the issue detail view for that issue

### Requirement: Issue detail view

The TUI SHALL display a breakdown of usage for a single issue. It SHALL show:
- Issue number, repo, and title
- A table of usage grouped by agent and model with columns: agent, model, calls, tokens in, tokens out, cost
- Total row: combined tokens, cost, and wall-clock time
- A navigation option to open the chart view for this issue

The user SHALL be able to filter the table by agent and/or model using filter inputs. Pressing Escape or Backspace at the top level SHALL return to the issue list.

#### Scenario: Detail view shows per-agent-model breakdown
- **WHEN** the user opens issue #42
- **WHEN** entries exist for two agents and two models
- **THEN** each unique (agent, model) pair is shown as a separate row

#### Scenario: Filter by agent
- **WHEN** the user applies a filter for agent "claude-code"
- **THEN** only rows with agent "claude-code" are shown
- **THEN** the total row updates to reflect the filtered subset

#### Scenario: Filter by model
- **WHEN** the user applies a filter for model "gpt-4o"
- **THEN** only rows with model "gpt-4o" are shown

### Requirement: Chart view

The TUI SHALL provide a chart view accessible from both the issue list (all-issues chart) and the issue detail view (single-issue chart). Charts SHALL render using `ntcharts` within the Bubble Tea model/update/view pattern.

The chart SHALL display token usage (tokens in + tokens out) over time. The user SHALL be able to toggle between day-level and week-level granularity. A date range picker SHALL allow filtering by start and end date. The user SHALL be able to filter by agent and/or model; the chart SHALL update to reflect the filtered data. A secondary display below the chart SHALL show total cost and total tokens for the selected range and filters.

#### Scenario: All-issues chart shows aggregated usage
- **WHEN** the user opens the chart from the issue list
- **THEN** the chart shows combined token usage across all issues grouped by day or week

#### Scenario: Single-issue chart shows usage for one issue
- **WHEN** the user opens the chart from issue detail view
- **THEN** the chart shows token usage for that issue only

#### Scenario: Day granularity
- **WHEN** the user selects day granularity
- **THEN** each bar or data point represents one calendar day

#### Scenario: Week granularity
- **WHEN** the user selects week granularity
- **THEN** each bar or data point represents one calendar week (Monday start)

#### Scenario: Date range filter
- **WHEN** the user sets a start date and end date
- **THEN** the chart shows only data within that range

#### Scenario: Agent filter on chart
- **WHEN** the user applies an agent filter
- **THEN** the chart reflects only token usage from that agent

#### Scenario: Model filter on chart
- **WHEN** the user applies a model filter
- **THEN** the chart reflects only token usage from that model

### Requirement: Keyboard navigation

The TUI SHALL support the following keyboard shortcuts:
- `q` or `Ctrl+C`: quit from any view
- Arrow keys / `j` / `k`: navigate lists
- `Enter`: select / drill down
- `Esc`: go back to previous view
- `c`: open chart view from issue list or detail view
- `f`: focus filter input
- `?`: show help overlay with all keybindings

#### Scenario: Help overlay
- **WHEN** the user presses `?`
- **THEN** a help overlay is shown listing all available keyboard shortcuts

#### Scenario: Quit from any view
- **WHEN** the user presses `q`
- **THEN** the TUI exits cleanly and returns to the shell
