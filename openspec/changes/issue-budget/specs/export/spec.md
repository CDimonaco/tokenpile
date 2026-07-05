## MODIFIED Requirements

### Requirement: Signed JSON export format

When exporting for a specific issue and a budget is set for that issue, the export document SHALL include a `budget` object at the top level with fields: `amount` (float), `spent` (float), `over` (bool). When no budget is set or when exporting across multiple issues, the `budget` field SHALL be omitted.

The `budget` object is NOT included in the signature computation. The signature remains over `entries` only.

#### Scenario: Export produces valid signed document
- **WHEN** `tokenpile export` is called
- **THEN** the output is valid JSON
- **THEN** the signature is valid

#### Scenario: Export includes budget when set
- **WHEN** `tokenpile export --issue 42` is called
- **WHEN** a budget of $5.00 is set for issue #42 and $3.20 has been spent
- **THEN** the document contains `"budget": {"amount": 5.00, "spent": 3.20, "over": false}`

#### Scenario: Export omits budget when not set
- **WHEN** `tokenpile export --issue 42` is called
- **WHEN** no budget is set for issue #42
- **THEN** the document does not contain a `budget` field

#### Scenario: Signature is over entries only
- **WHEN** two exports are produced with the same entries but different budget amounts
- **THEN** both exports have the same signature
