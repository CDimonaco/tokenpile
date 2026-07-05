### Requirement: Budget set and unset commands

The system SHALL provide `tokenpile budget set --repo owner/repo --issue N --amount X.XX` to store a budget for an issue and `tokenpile budget unset --repo owner/repo --issue N` to remove it. The `--repo` flag SHALL be inferred from git remote when absent.

#### Scenario: Set a budget
- **WHEN** `tokenpile budget set --issue 42 --amount 5.00`
- **THEN** a budget of $5.00 is stored for that issue
- **THEN** the CLI prints `Budget set: cdimonaco/repo #42 → $5.00` and exits 0

#### Scenario: Overwrite an existing budget
- **WHEN** a budget of $5.00 already exists for issue #42
- **WHEN** `tokenpile budget set --issue 42 --amount 10.00`
- **THEN** the budget is updated to $10.00

#### Scenario: Unset a budget
- **WHEN** `tokenpile budget unset --issue 42`
- **THEN** the budget for that issue is removed
- **THEN** the CLI prints `Budget removed: cdimonaco/repo #42` and exits 0

#### Scenario: Unset a budget that does not exist
- **WHEN** `tokenpile budget unset --issue 42` is called and no budget exists
- **THEN** the CLI exits 0 with no error (idempotent)

#### Scenario: Invalid amount
- **WHEN** `tokenpile budget set --issue 42 --amount -1`
- **THEN** the CLI exits with a non-zero code and an error: "amount must be greater than zero"

### Requirement: Budget persistence

Budgets SHALL be stored in the local SQLite DB in an `issue_budgets` table keyed by `(repo, issue_num)`. Budgets are local to the machine and are not exported as signed data.

#### Scenario: Budget survives process restart
- **WHEN** a budget is set for issue #42
- **WHEN** the process exits and restarts
- **THEN** the budget is still retrievable from the DB
