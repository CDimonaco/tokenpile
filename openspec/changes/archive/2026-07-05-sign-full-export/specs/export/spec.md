## MODIFIED Requirements

### Requirement: Signed JSON export format

Export documents SHALL carry `schema_version: "3.0"`. The Ed25519 signature SHALL be computed over the SHA-256 digest of the canonical JSON (recursively key-sorted objects, no insignificant whitespace) of the entire document with the `signature` field set to the empty string. All fields — `schema_version`, `exported_at`, `exported_by`, `public_key`, `entries`, `sessions`, `budgets` — are inside the signed surface.

When exporting for a specific issue and a budget is set for that issue, the export document SHALL include the budget data. When no budget is set or when exporting across multiple issues, budget data SHALL be omitted.

#### Scenario: Export produces valid signed document
- **WHEN** `tokenpile export` is called
- **THEN** the output is valid JSON with `schema_version` equal to `3.0`
- **THEN** the signature verifies against the full document minus the `signature` field

#### Scenario: Export includes budget when set
- **WHEN** `tokenpile export --issue 42` is called
- **WHEN** a budget of $5.00 is set for issue #42
- **THEN** the document contains the budget amount for issue #42

#### Scenario: Export omits budget when not set
- **WHEN** `tokenpile export --issue 42` is called
- **WHEN** no budget is set for issue #42
- **THEN** the document does not contain budget data

#### Scenario: Any field tampering invalidates the signature
- **WHEN** a 3.0 export file is modified in any signed field (an entry, a session note, a budget amount, or `exported_at`)
- **THEN** `tokenpile export verify --file <path>` reports the signature as invalid and exits non-zero

## ADDED Requirements

### Requirement: Verification of legacy 2.0 exports

`tokenpile export verify` SHALL detect documents with `schema_version: "2.0"` and verify them with the legacy rule (signature over the canonical JSON of `entries` only). When a 2.0 document verifies successfully, the command SHALL print a warning stating that sessions and budgets are not covered by the signature. Documents with any other `schema_version` SHALL fail verification with an unsupported-version error.

#### Scenario: Legacy file verifies with warning
- **WHEN** `tokenpile export verify --file old.json` is called on a valid 2.0 export
- **THEN** verification succeeds
- **THEN** the output warns that only entries are covered by the signature

#### Scenario: Unknown version rejected
- **WHEN** `tokenpile export verify --file doc.json` is called on a document with `schema_version: "9.9"`
- **THEN** verification fails with an unsupported schema version error

### Requirement: Origin verification with expected public key

`tokenpile export verify` SHALL accept an optional `--pubkey` flag whose value is either a base64-encoded Ed25519 public key or a path to a file containing one (PEM `ED25519 PUBLIC KEY` block or base64 text). When provided, verification SHALL additionally require the document's embedded public key to equal the expected key, compared in constant time. When not provided, the output SHALL state that the check proves internal consistency only.

#### Scenario: Matching pubkey proves origin
- **WHEN** `tokenpile export verify --file doc.json --pubkey <key>` is called and the document key matches
- **THEN** verification succeeds and the output states the origin was verified

#### Scenario: Mismatched pubkey fails
- **WHEN** `tokenpile export verify --file doc.json --pubkey <key>` is called and the document key differs
- **THEN** the command exits non-zero reporting a public key mismatch, even if the signature is internally valid

#### Scenario: Without pubkey the guarantee is labeled
- **WHEN** `tokenpile export verify --file doc.json` is called without `--pubkey` on a valid document
- **THEN** the output states the signature was checked against the embedded key (consistency only)
