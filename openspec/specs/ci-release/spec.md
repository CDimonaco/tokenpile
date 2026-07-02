
### Requirement: asdf-vm version pinning

The repository SHALL include a `.tool-versions` file at the project root pinning the versions of all dev dependencies: `golang`, `golangci-lint`, `goreleaser`, and `mockery`. This file SHALL be the single source of truth for tool versions used by both local development and CI.

#### Scenario: Developer sets up environment
- **WHEN** a developer runs `asdf install` in the repository root
- **THEN** all pinned tool versions are installed and available in PATH

#### Scenario: CI uses same versions as local
- **WHEN** the CI workflow runs
- **THEN** it reads `.tool-versions` via the asdf GitHub Action and installs the same tool versions used locally

### Requirement: Makefile as task runner

The repository SHALL include a `Makefile` at the project root providing the following targets: `build`, `test`, `lint`, `fmt`, `generate`, `install`, `clean`, `release-check`. All CI steps SHALL invoke make targets rather than raw tool commands to keep local and CI behavior identical.

#### Scenario: make test runs tests with race detector
- **WHEN** a developer runs `make test`
- **THEN** `go test -race ./...` is executed

#### Scenario: make release-check validates goreleaser config
- **WHEN** a developer runs `make release-check`
- **THEN** `goreleaser check` is executed and exits non-zero if the config is invalid

#### Scenario: make generate regenerates mocks
- **WHEN** a developer runs `make generate`
- **THEN** mockery regenerates all mock files from their source interfaces

### Requirement: CI workflow on every push

The repository SHALL include a GitHub Actions workflow at `.github/workflows/ci.yml` that runs on every push and pull request to any branch. The workflow SHALL install tool versions from `.tool-versions` using the asdf GitHub Action, then execute in sequence via make targets:
1. `make lint`
2. `make fmt` (fail if any files are not formatted)
3. `make test`

#### Scenario: CI passes on clean code
- **WHEN** a push is made with code that passes lint, is formatted, and all tests pass
- **THEN** the CI workflow completes successfully

#### Scenario: CI fails on lint error
- **WHEN** a push introduces a lint violation
- **THEN** the golangci-lint step fails and the workflow stops

#### Scenario: CI fails on unformatted code
- **WHEN** a push includes unformatted Go files
- **THEN** the gofmt step fails

#### Scenario: CI fails on test failure
- **WHEN** a push introduces a failing test
- **THEN** the test step fails

### Requirement: Release workflow on version tag

The repository SHALL include a GitHub Actions workflow at `.github/workflows/release.yml` that triggers on tags matching `v*.*.*`. The workflow SHALL use goreleaser to:
1. Build binaries for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
2. Create a GitHub Release with the built binaries and a `checksums.txt`
3. Push an updated Homebrew formula to the `cdimonaco/homebrew-tokenpile` repository

#### Scenario: Release creates binaries for all platforms
- **WHEN** a tag `v0.1.0` is pushed
- **THEN** goreleaser builds four platform binaries
- **THEN** a GitHub Release is created with all binaries and checksums attached

#### Scenario: Release updates Homebrew tap
- **WHEN** a tag `v0.1.0` is pushed
- **THEN** the formula in `cdimonaco/homebrew-tokenpile` is updated to point to the new release

### Requirement: goreleaser configuration

The repository SHALL include a `.goreleaser.yaml` at the project root defining:
- Build targets: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`
- Archive format: `.tar.gz`
- Checksum file: `checksums.txt` (SHA256)
- Homebrew tap config: repo `cdimonaco/homebrew-tokenpile`, formula name `tokenpile`
- CGO disabled (`CGO_ENABLED=0`) for all targets

#### Scenario: goreleaser config is valid
- **WHEN** `goreleaser check` is run
- **THEN** the config passes validation without errors

### Requirement: golangci-lint configuration

The repository SHALL include a `.golangci.yml` at the project root using the maratori golden configuration as the base. The linter SHALL be run with `--timeout 5m` in CI.

#### Scenario: golangci-lint runs with project config
- **WHEN** `golangci-lint run` is executed in the repository root
- **THEN** it picks up `.golangci.yml` and applies the maratori golden config rules

### Requirement: Homebrew tap repository

A separate public GitHub repository `cdimonaco/homebrew-tokenpile` SHALL exist and be configured as the goreleaser Homebrew tap target. Installation via Homebrew SHALL work as:
```
brew tap cdimonaco/tokenpile
brew install tokenpile
```

#### Scenario: Homebrew install works after release
- **WHEN** a release has been published and the tap updated
- **WHEN** a macOS user runs `brew tap cdimonaco/tokenpile && brew install tokenpile`
- **THEN** the `tokenpile` binary is installed and available in PATH
