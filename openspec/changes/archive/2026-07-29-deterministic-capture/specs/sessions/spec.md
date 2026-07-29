## MODIFIED Requirements

### Requirement: Session domain type

The system SHALL define a `Session` type in `internal/usage/` with:
- `ID` string (UUID)
- `Repo` string
- `IssueNum` *int (nil when the session is not attributed to an issue)
- `StartedAt` time.Time (UTC)
- `EndedAt` *time.Time (nil if active)
- `Note` string (empty if not set)
- `Tags` []string (empty slice if not set)

A session's issue MAY be unset at creation and assigned later. Capture must be able to open a session before the issue is known, so attribution is an annotation rather than a precondition.

#### Scenario: Active session has nil EndedAt
- **WHEN** a session has not been closed
- **THEN** `EndedAt` is nil

#### Scenario: Closed session has non-nil EndedAt
- **WHEN** a session has been closed
- **THEN** `EndedAt` is the time it was closed

#### Scenario: Session with no annotations has empty note and tags
- **WHEN** a session was created without any `--note` or `--tag`
- **THEN** `Note` is an empty string
- **THEN** `Tags` is an empty slice (not nil)

#### Scenario: Session without an issue
- **WHEN** a session is created by capture with no issue resolved
- **THEN** `IssueNum` is nil
- **THEN** the session records usage normally

#### Scenario: Issue assigned after the fact
- **WHEN** a session with `IssueNum` nil is assigned to issue 42
- **THEN** `IssueNum` is 42
- **THEN** its recorded usage is unchanged
