## MODIFIED Requirements

### Requirement: Issue detail view

The detail view SHALL display a two-tab layout: **Summary** and **Sessions**. The active tab SHALL be indicated visually. Pressing `tab` SHALL cycle between the two tabs.

The **Summary** tab SHALL show the existing per-agent-model aggregated breakdown (unchanged behavior).

The **Sessions** tab SHALL list all sessions for the issue in chronological order. Each session row SHALL show:
- Start time and end time (or "active" if ongoing)
- Duration
- Tag badges (empty if no tags)
- Note text on a second line (omitted if empty)
- Token and cost totals for all entries in that session

#### Scenario: Detail view shows per-agent-model breakdown
- **WHEN** the user navigates to an issue detail
- **THEN** the Summary tab is active by default
- **THEN** per-agent-model rows are shown as before

#### Scenario: Filter by agent
- **WHEN** the user presses `a` in the Summary tab
- **THEN** usage is filtered to the selected agent

#### Scenario: Filter by model
- **WHEN** the user presses `m` in the Summary tab
- **THEN** usage is filtered to the selected model

#### Scenario: Tab switches to Sessions view
- **WHEN** the user presses `tab` in the detail view
- **THEN** the Sessions tab becomes active
- **THEN** sessions are listed chronologically with note and tag badges

#### Scenario: Session with note and tags displayed
- **WHEN** a session has note "fixed unicode handling" and tags ["refactor", "bug"]
- **THEN** the note text appears below the session row
- **THEN** "refactor" and "bug" appear as badges on the session row

#### Scenario: Session without annotations shown cleanly
- **WHEN** a session has no note and no tags
- **THEN** no note line is shown
- **THEN** no tag badges are shown

## ADDED Requirements

### Requirement: Sessions tab keyboard navigation

The Sessions tab SHALL support scrolling through the session list with arrow keys when the list exceeds the terminal height.

#### Scenario: Scroll through sessions
- **WHEN** the Sessions tab is active and there are more sessions than fit on screen
- **WHEN** the user presses the down arrow
- **THEN** the list scrolls to reveal the next session
