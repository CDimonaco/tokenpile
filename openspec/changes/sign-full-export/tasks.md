## 1. Export package: full-document signing

- [x] 1.1 Change `SchemaVersion` to `"3.0"` in `internal/export/export.go`
- [x] 1.2 Rework `Build` to construct the complete `Document` with `Signature: ""`, canonicalize it, sign the SHA-256 digest, then set the `Signature` field
- [x] 1.3 Rework `Verify` to dispatch on `schema_version`: `"3.0"` verifies the full document with `signature` blanked; `"2.0"` verifies the legacy entries-only rule and returns a distinct legacy result the caller can surface; any other version returns an unsupported-version error
- [x] 1.4 Add a pinned-canonical-bytes fixture test so a future Go version changing float formatting fails CI
- [x] 1.5 Update `internal/export/export_test.go`: tamper tests for sessions, budgets and `exported_at` on 3.0 documents; committed 2.0 fixture verifying with the legacy path

## 2. Schema

- [x] 2.1 Update the embedded JSON schema in `internal/schema` for version 3.0 if it encodes the version or signature semantics

## 3. CLI: verify command

- [x] 3.1 Add `--pubkey` flag to `export verify` in `cmd/tokenpile/cmd_export.go`, accepting base64 or a file path (PEM `ED25519 PUBLIC KEY` block or base64 text)
- [x] 3.2 Compare the expected key with the document key using `crypto/subtle` before signature verification; fail non-zero on mismatch
- [x] 3.3 Update `verify` output: schema version, signature scope (full document vs legacy entries-only warning), origin check status (verified via `--pubkey` / consistency only)
- [x] 3.4 Integration tests in `cmd/tokenpile/integration_test.go`: 3.0 roundtrip, tampered session fails, legacy 2.0 warns, pubkey match and mismatch

## 4. Docs

- [x] 4.1 Update CLAUDE.md key design decisions with the new signature scope and the origin-verification option
- [x] 4.2 Run `make check` and commit
