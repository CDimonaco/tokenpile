### Requirement: Reset command with confirmation

The system SHALL provide `tokenpile reset` which removes all local tokenpile state: the SQLite database and its WAL/SHM sidecar files, the Ed25519 identity keypair, the encrypted credentials file, the keychain token, the pricing override file, and installed agent skills. Before doing anything, the command SHALL print the list of items that will be deleted and require the user to type `yes`; any other input SHALL abort with a non-zero exit code and no side effects. A `--yes` flag SHALL skip the prompt for non-interactive use.

#### Scenario: Confirmation aborts on anything but yes
- **WHEN** the user runs `tokenpile reset` and types `no` at the prompt
- **THEN** the command exits non-zero
- **THEN** no file is deleted and no backup is written

#### Scenario: Non-interactive reset
- **WHEN** the user runs `tokenpile reset --yes`
- **THEN** no prompt is shown and the reset proceeds

#### Scenario: Full state removal
- **WHEN** a confirmed reset completes
- **THEN** the database, identity keypair, credentials file, pricing override and installed agent skills no longer exist
- **THEN** the keychain no longer holds a tokenpile token
- **THEN** the command prints each removed item

### Requirement: Signed backup before destruction

Unless `--no-backup` is passed, `tokenpile reset` SHALL write a signed export document (the standard export format) containing all entries, all sessions and all budgets before deleting anything. The dump SHALL be signed with the identity key prior to its deletion. The default path is `tokenpile-backup-<timestamp>.json` in the working directory; `--output` overrides it. If writing the backup fails, the command SHALL abort without deleting anything. The dump format is the standard export document so a future import command can restore it.

#### Scenario: Backup written and verifiable
- **WHEN** the user runs `tokenpile reset --yes`
- **THEN** a backup file exists at the default path
- **THEN** `tokenpile export verify --file <backup>` succeeds against the embedded key

#### Scenario: Backup failure aborts reset
- **WHEN** the backup cannot be written (e.g. the output directory does not exist)
- **THEN** the command exits non-zero and no state is deleted

#### Scenario: Skipping the backup
- **WHEN** the user runs `tokenpile reset --yes --no-backup`
- **THEN** no backup file is written and the reset proceeds

### Requirement: Partial failure reporting and idempotence

When a deletion step fails after a successful backup, `tokenpile reset` SHALL continue with the remaining deletions, report every failed item, and exit non-zero. Running `tokenpile reset` when some or all state is already absent SHALL succeed, treating missing items as already removed.

#### Scenario: Rerun after partial failure completes the job
- **WHEN** `tokenpile reset --yes` is run twice in a row
- **THEN** the second run succeeds reporting nothing left to remove (beyond a fresh backup, if not skipped)
