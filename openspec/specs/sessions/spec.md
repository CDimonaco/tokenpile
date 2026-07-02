
### Requirement: Implicit session start on first log

The system SHALL automatically start a new session for an issue when `tokenpile log` is called and no active session exists for that `(repo, issue_num)` pair. The agent and user SHALL NOT be required to call any session management command.

#### Scenario: Session starts on first log
- **WHEN** `tokenpile log` is called for issue #42
- **WHEN** no active session exists for `(repo, 42)`
- **THEN** a new session is created with `started_at = now()`
- **THEN** the `UsageEntry` references the new session ID

#### Scenario: Subsequent logs reuse active session
- **WHEN** `tokenpile log` is called for issue #42
- **WHEN** an active session already exists for `(repo, 42)`
- **THEN** the existing session ID is used for the new `UsageEntry`
- **THEN** no new session is created

### Requirement: Session idle auto-close

The system SHALL automatically close any session that has had no `UsageEntry` associated with it for 30 minutes or more. The check SHALL occur each time `tokenpile log` is called: before starting or reusing a session, the system SHALL close all sessions for that issue that have been idle beyond the threshold.

#### Scenario: Idle session closed before new log
- **WHEN** `tokenpile log` is called for issue #42
- **WHEN** the most recent `UsageEntry` for the active session was logged more than 30 minutes ago
- **THEN** the active session is closed (`ended_at = now()`)
- **THEN** a new session is started

#### Scenario: Active session not closed when within threshold
- **WHEN** `tokenpile log` is called for issue #42
- **WHEN** the most recent `UsageEntry` for the active session was logged less than 30 minutes ago
- **THEN** the active session remains open

### Requirement: Wall-clock time aggregation

The system SHALL compute total wall-clock time for an issue by summing the durations of all closed sessions: `sum(ended_at - started_at)`. Active (unclosed) sessions SHALL contribute `now() - started_at` to the running total in reports.

#### Scenario: Time reported for issue with multiple sessions
- **WHEN** issue #42 has two closed sessions of 35 minutes and 12 minutes
- **WHEN** `tokenpile report --issue 42` is called
- **THEN** the reported time is 47 minutes

#### Scenario: Time reported includes active session
- **WHEN** issue #42 has one closed session of 20 minutes and one active session started 10 minutes ago
- **WHEN** `tokenpile report --issue 42` is called
- **THEN** the reported time is approximately 30 minutes

### Requirement: Session domain type

The system SHALL define a `Session` type in `internal/domain/` with:
- `ID` string (UUID)
- `Repo` string
- `IssueNum` int
- `StartedAt` time.Time (UTC)
- `EndedAt` *time.Time (nil if active)

#### Scenario: Active session has nil EndedAt
- **WHEN** a session has not been closed
- **THEN** `EndedAt` is nil

#### Scenario: Closed session has non-nil EndedAt
- **WHEN** a session has been closed
- **THEN** `EndedAt` is the time it was closed
