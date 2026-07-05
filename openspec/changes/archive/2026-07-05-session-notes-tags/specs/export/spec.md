## MODIFIED Requirements

### Requirement: Signed JSON export format

The export document SHALL have `schema_version: "2.0"`. It SHALL contain:
- `entries`: array of usage entries (unchanged structure, signature still covers only this array)
- `sessions`: array of session objects, each with `id`, `repo`, `issue_num`, `started_at`, `ended_at` (nullable), `note` (omitted if empty), `tags` (omitted if empty)

The signature SHALL remain over the canonical JSON of `entries` only. Sessions are supplementary context and are not signed.

#### Scenario: Export produces valid signed document
- **WHEN** `tokenpile export` is called
- **THEN** the output is valid JSON conforming to the v2 schema
- **THEN** `schema_version` is `"2.0"`
- **THEN** `entries` and `sessions` arrays are present

#### Scenario: Signature is over entries only
- **WHEN** two exports are produced with the same entries but different session notes
- **THEN** both exports have the same signature

#### Scenario: Sessions included in export
- **WHEN** an issue has sessions with notes and tags
- **WHEN** `tokenpile export --issue 42` is called
- **THEN** the `sessions` array contains those sessions with `note` and `tags` fields

#### Scenario: Session without annotations omits empty fields
- **WHEN** a session has no note and no tags
- **THEN** the session object in the export omits the `note` and `tags` fields

### Requirement: Export JSON Schema

The system SHALL embed a JSON Schema that validates the v2 export format, including the `sessions` array structure.

#### Scenario: Schema embedded in binary
- **WHEN** the binary is built
- **THEN** the JSON schema for v2 is accessible at runtime without external files

### Requirement: Export verify command

The system SHALL verify the Ed25519 signature over the `entries` array. Verification SHALL succeed regardless of whether the `sessions` array is present or modified.

#### Scenario: Valid file
- **WHEN** `tokenpile export verify --file export.json` is called with an unmodified export
- **THEN** the CLI prints `OK: signature valid, N entries` and exits 0

#### Scenario: Tampered entries
- **WHEN** an entry in the `entries` array is modified after export
- **THEN** verification fails with a non-zero exit code

#### Scenario: Schema validation failure
- **WHEN** the file does not conform to the v2 JSON schema
- **THEN** verification fails with a descriptive error
