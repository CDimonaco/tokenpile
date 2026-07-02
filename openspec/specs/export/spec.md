
### Requirement: Export command with scope and filter flags

The system SHALL provide a `tokenpile export` command that writes a signed JSON file to stdout or a specified output path. The command SHALL support the following flags:
- `--issue <num>` (optional): restrict to a single issue
- `--repo <owner/repo>` (optional, inferred): restrict to a repo
- `--from <date>` (optional): start of date range (RFC3339 date)
- `--to <date>` (optional): end of date range (RFC3339 date)
- `--agent <name>` (optional): restrict to a specific agent
- `--model <name>` (optional): restrict to a specific model
- `--output <path>` (optional): write to file instead of stdout

All flags are combinable. With no flags, all data is exported.

#### Scenario: Export all data
- **WHEN** `tokenpile export` is run with no flags
- **THEN** a signed JSON containing all usage entries is written to stdout

#### Scenario: Export single issue
- **WHEN** `tokenpile export --issue 42 --repo owner/repo`
- **THEN** only entries for issue #42 in that repo are included

#### Scenario: Export with date range
- **WHEN** `tokenpile export --from 2026-06-01 --to 2026-06-30`
- **THEN** only entries with `at` within the range are included

#### Scenario: Export to file
- **WHEN** `tokenpile export --output usage.json`
- **THEN** the signed JSON is written to `usage.json`
- **THEN** the CLI prints "Exported to usage.json"

### Requirement: Signed JSON export format

The exported file SHALL be a self-contained JSON document conforming to a published JSON Schema. The document SHALL include:
- `schema_version`: string (e.g., `"1.0"`)
- `exported_at`: RFC3339 timestamp
- `exported_by`: string (e.g., `"tokenpile/0.1.0"`)
- `public_key`: base64-encoded Ed25519 public key
- `entries`: array of usage entry objects (see `UsageEntry` domain type)
- `signature`: base64-encoded Ed25519 signature over the canonical JSON of `entries`

The signature SHALL be computed over `sha256(canonical_json(entries))` where canonical JSON is RFC 8785 (deterministic key ordering, no extra whitespace).

#### Scenario: Export produces valid signed document
- **WHEN** `tokenpile export` is run
- **THEN** the output is valid JSON matching the export JSON Schema
- **THEN** the signature field is present and non-empty

#### Scenario: Signature is over entries only
- **WHEN** the `entries` array changes
- **THEN** the signature changes
- **WHEN** only metadata fields change (e.g., `exported_at`)
- **THEN** the signature remains valid for the same entries

### Requirement: Export JSON Schema

The system SHALL maintain a JSON Schema (draft 2020-12) for the export format at `schema/export.schema.json` in the repository. This schema SHALL be embedded in the binary and used for validation on import and verify operations.

#### Scenario: Schema embedded in binary
- **WHEN** the binary is built
- **THEN** `schema/export.schema.json` is accessible at runtime without external files

### Requirement: Export verify command

The system SHALL provide `tokenpile export verify --file <path>` which:
1. Reads and parses the JSON file
2. Validates it against the embedded JSON Schema
3. Verifies the Ed25519 signature using the public key embedded in the file
4. Prints a clear result: valid, schema invalid, or signature mismatch

#### Scenario: Valid file
- **WHEN** `tokenpile export verify --file usage.json`
- **WHEN** the file is unmodified since export
- **THEN** the CLI prints "Signature valid. Exported by <public_key_fingerprint>"

#### Scenario: Tampered entries
- **WHEN** the `entries` array in the file has been modified
- **THEN** the CLI prints "Signature invalid: entries have been tampered with" and exits non-zero

#### Scenario: Schema validation failure
- **WHEN** the file does not conform to the JSON Schema
- **THEN** the CLI prints the schema validation errors and exits non-zero

### Requirement: Per-machine signing identity

Each tokenpile installation has its own Ed25519 keypair (per the `auth` spec). Export files are signed with the local private key and include the corresponding public key. Recipients verifying the file can confirm integrity but cannot verify the signer's real-world identity without out-of-band public key exchange.

#### Scenario: Export from two machines produces different public keys
- **WHEN** two machines each have their own `identity.key`
- **THEN** exports from each machine embed different public keys and different signatures
