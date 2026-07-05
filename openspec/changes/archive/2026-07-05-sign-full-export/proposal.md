## Why

The export signature currently covers only the `entries` array — a deliberate decision in the current spec ("The signature remains over `entries` only"). In practice this makes `tokenpile export verify` misleading: sessions, budgets, `exported_at`, and `schema_version` can be tampered with and the file still verifies as "OK: signature valid". Verification also trusts the public key embedded in the document itself, so it proves internal consistency only, never origin. As exports gain fields (sessions, budgets), the unsigned surface keeps growing.

## What Changes

- **BREAKING** The Ed25519 signature SHALL cover the canonical JSON of the entire export document except the `signature` field itself (entries, sessions, budgets, `exported_at`, `exported_by`, `public_key`, `schema_version`).
- **BREAKING** `schema_version` bumps from `2.0` to `3.0`.
- `tokenpile export verify` gains an optional `--pubkey <base64|path>` flag: when given, verification additionally requires the document's embedded public key to match the expected key, proving origin and not just consistency.
- `verify` output SHALL state what the signature covers and whether origin was checked (embedded key vs `--pubkey`).
- Compatibility: `verify` SHALL detect `schema_version: "2.0"` documents and verify them with the legacy entries-only rule, printing a warning that sessions/budgets are not covered by the signature. Export always writes 3.0.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `export`: the "Signed JSON export format" requirement changes — signature scope extends from entries-only to the full document minus `signature`; version bump to 3.0; new `--pubkey` verification option; legacy 2.0 verification behavior defined.

## Impact

- `internal/export/export.go`: `Build` signs the whole document; `Verify` dispatches on `schema_version` (3.0 full-document, 2.0 legacy entries-only); `SchemaVersion` constant to `3.0`.
- `internal/schema` (embedded JSON schema): update for 3.0 if version-specific.
- `cmd/tokenpile/cmd_export.go`: `--pubkey` flag on `verify`, updated output messages.
- Tests: `internal/export/export_test.go`, `cmd/tokenpile/integration_test.go` (tamper scenarios for sessions/budgets, legacy 2.0 fixture, pubkey match/mismatch).
- Docs: CLAUDE.md key design decisions (signature scope).
- Existing 2.0 export files remain verifiable (legacy path); any external consumer parsing exports must handle the new version string.
