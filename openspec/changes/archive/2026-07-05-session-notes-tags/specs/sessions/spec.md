## MODIFIED Requirements

### Requirement: Session domain type

The system SHALL define a `Session` type in `internal/usage/` with:
- `ID` string (UUID)
- `Repo` string
- `IssueNum` int
- `StartedAt` time.Time (UTC)
- `EndedAt` *time.Time (nil if active)
- `Note` string (empty if not set)
- `Tags` []string (empty slice if not set)

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
