## MODIFIED Requirements

### Requirement: Signed JSON export format

Export documents SHALL carry `schema_version: "4.0"`. The Ed25519 signature SHALL be computed over the SHA-256 digest of the canonical JSON (recursively key-sorted objects, no insignificant whitespace) of the entire document with the `signature` field set to the empty string. All fields — `schema_version`, `exported_at`, `exported_by`, `public_key`, `entries`, `sessions`, `budgets` — are inside the signed surface.

Each entry SHALL carry its token counts split by billing tier (`input_fresh`, `cache_write`, `cache_read`, `output`, `reasoning`) and a `source` of `measured` or `estimated`. Entries SHALL NOT carry aggregate `tokens_in` or `tokens_out` fields: a consumer that wants a total computes it from the tiers, so no field exists whose meaning depends on which tiers were populated.

Sessions and budgets SHALL follow the same repo/issue scope as the entries filter: an unfiltered export includes all sessions and all budgets; an export filtered by `--repo` includes that repository's sessions and budgets; an export filtered by `--repo` and `--issue` includes only that issue's. Time, agent and model filters do not restrict sessions or budgets. When no budget exists in the selected scope, budget data SHALL be omitted.

#### Scenario: Export produces valid signed document
- **WHEN** `tokenpile export` is called
- **THEN** the output is valid JSON with `schema_version` equal to `4.0`
- **THEN** the signature verifies against the full document minus the `signature` field

#### Scenario: Entries carry tiers and provenance
- **WHEN** an export contains an entry logged with fresh input and output tokens
- **THEN** that entry object carries all five tier fields and a `source`
- **THEN** it carries no `tokens_in` or `tokens_out` field

#### Scenario: Unfiltered export includes all sessions and budgets
- **WHEN** usage exists for issues #1 and #2 with sessions, and a budget is set for issue #2
- **WHEN** `tokenpile export` is called without filters
- **THEN** the document contains the sessions of both issues and the budget of issue #2

#### Scenario: Repo-filtered export scopes sessions and budgets to the repo
- **WHEN** usage exists in `owner/a` and `owner/b`
- **WHEN** `tokenpile export --repo owner/a` is called
- **THEN** the document contains only sessions and budgets belonging to `owner/a`

#### Scenario: Export includes budget when set
- **WHEN** `tokenpile export --issue 42 --repo owner/repo` is called
- **WHEN** a budget of $5.00 is set for issue #42
- **THEN** the document contains the budget amount for issue #42

#### Scenario: Export omits budget when not set
- **WHEN** `tokenpile export --issue 42 --repo owner/repo` is called
- **WHEN** no budget is set for issue #42
- **THEN** the document does not contain budget data

#### Scenario: Any field tampering invalidates the signature
- **WHEN** a 4.0 export file is modified in any signed field (an entry tier, a session note, a budget amount, or `exported_at`)
- **THEN** `tokenpile export verify --file <path>` reports the signature as invalid and exits non-zero

## ADDED Requirements

### Requirement: Verification of pre-4.0 exports

`tokenpile export verify` SHALL continue to verify documents carrying `schema_version` `2.0` and `3.0` under the rules those versions were written with, reporting the version it applied. Older documents exist on disk — including `reset` backups — and the verification code for them is already implemented and tested.

Writing pre-4.0 documents SHALL NOT be supported: `tokenpile export` always emits 4.0.

#### Scenario: A 3.0 document still verifies
- **WHEN** `tokenpile export verify --file <3.0 document>` is called
- **THEN** verification succeeds and the output states the document is schema 3.0

#### Scenario: Export never writes an old version
- **WHEN** `tokenpile export` is called
- **THEN** the document carries `schema_version` `4.0` with no option to emit an earlier version
