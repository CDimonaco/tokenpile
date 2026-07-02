
### Requirement: IssueProvider interface

The system SHALL define an `IssueProvider` interface in `internal/provider/` that abstracts issue fetching. The interface SHALL expose:
- `ListIssues(ctx context.Context, filter IssueFilter) ([]Issue, error)`
- `GetIssue(ctx context.Context, repo string, number int) (*Issue, error)`

`IssueFilter` SHALL support filtering by: repo, state (open/closed/all), assignee.

#### Scenario: Interface satisfied by GitHub implementation
- **WHEN** a `GitHubIssueProvider` is constructed with a valid token
- **THEN** it satisfies the `IssueProvider` interface at compile time

### Requirement: GitHub issue provider implementation

The system SHALL implement `IssueProvider` using the GitHub REST API via `google/go-github`. The implementation SHALL request the minimum OAuth scopes required: `read:user` and `repo` (for private repositories).

#### Scenario: List issues for a repository
- **WHEN** `ListIssues` is called with a valid repo and filter
- **THEN** it returns all matching issues from the GitHub API
- **THEN** each issue includes: number, title, state, URL, and repo identifier

#### Scenario: Get a specific issue
- **WHEN** `GetIssue` is called with a valid repo and issue number
- **THEN** it returns the issue or a typed error if not found

#### Scenario: Unauthenticated request
- **WHEN** `ListIssues` or `GetIssue` is called without a valid token
- **THEN** it returns a sentinel error `ErrUnauthenticated`
- **THEN** the caller MAY prompt the user to run `tokenpile auth login`

### Requirement: Repo identifier format

The system SHALL use the `owner/repo` string format as the canonical repo identifier throughout all CLI flags, storage, and display. The system SHALL validate this format when provided explicitly and reject values that do not match.

#### Scenario: Valid repo identifier
- **WHEN** `--repo cdimonaco/tokenpile` is passed
- **THEN** it is accepted and used as-is

#### Scenario: Invalid repo identifier
- **WHEN** `--repo tokenpile` is passed (missing owner)
- **THEN** the CLI exits with an error: "invalid repo format, expected owner/repo"

### Requirement: Repo inference from git remote

When `--repo` is not provided, the system SHALL attempt to infer the repository from the current directory's git remote. It SHALL run `git remote get-url origin` and parse the result into `owner/repo` format. Both HTTPS (`https://github.com/owner/repo.git`) and SSH (`git@github.com:owner/repo.git`) formats SHALL be supported.

#### Scenario: Successful inference from HTTPS remote
- **WHEN** `--repo` is not provided
- **WHEN** `git remote get-url origin` returns `https://github.com/owner/repo.git`
- **THEN** `owner/repo` is used as the repo identifier

#### Scenario: Successful inference from SSH remote
- **WHEN** `--repo` is not provided
- **WHEN** `git remote get-url origin` returns `git@github.com:owner/repo.git`
- **THEN** `owner/repo` is used as the repo identifier

#### Scenario: Inference fails — no git remote
- **WHEN** `--repo` is not provided
- **WHEN** the current directory has no git remote or is not a git repository
- **THEN** the CLI exits with an error: "cannot infer repo: not a git repository or no origin remote configured; pass --repo owner/repo"

#### Scenario: Inference fails — non-GitHub remote
- **WHEN** `--repo` is not provided
- **WHEN** the origin remote points to a non-GitHub host
- **THEN** the CLI exits with an error indicating inference failed and instructing the user to pass `--repo`
