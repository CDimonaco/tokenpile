## Context

`internal/export.Build` marshals entries to canonical JSON (sorted object keys, recursive), hashes with SHA-256, and signs the digest with the user's Ed25519 identity key. `Verify` re-derives the canonical JSON of `doc.Entries` and checks the signature against the public key embedded in the document. The current spec explicitly scopes the signature to entries only; sessions and budgets were added later outside the signed surface. `SchemaVersion` is `"2.0"` and is emitted in every document.

Constraints: no new dependencies; the identity key format and location do not change; exports must remain a single self-contained JSON file.

## Goals / Non-Goals

**Goals:**
- Tampering with any field of an export (except whitespace/key order) fails verification.
- `verify` can prove origin against a caller-supplied expected public key.
- Existing 2.0 files keep verifying, with an explicit warning about their weaker guarantee.

**Non-Goals:**
- Key distribution or trust infrastructure (how the verifier obtains the expected pubkey is out of scope; it is passed explicitly).
- Signing anything other than exports (DB rows stay unsigned; cost is still computed at report time).
- Re-signing or upgrading existing 2.0 files.

## Decisions

**1. Sign the full document with `signature` emptied, not a detached sub-document.**
Serialize the `Document` with `Signature: ""`, canonicalize, hash, sign, then set the field. Verify reverses this: copy the document, blank `signature`, canonicalize, check. Alternative considered: signing a concatenation of per-section digests (entries, sessions, budgets) — rejected as more code for no gain; the whole-document rule is simpler to state in the spec and leaves no field accidentally outside the signed surface, including future ones.

**2. Keep the existing canonical JSON implementation.**
`canonicalJSON` (marshal, re-unmarshal to `any`, recursive key-sorted emit) already provides deterministic output and is covered by tests. JCS (RFC 8785) compliance is not needed since only this tool produces and verifies these documents. Risk of float round-tripping (`budget.amount` is a float64) is acceptable: Go's `encoding/json` round-trips float64 deterministically on the same codebase.

**3. Version dispatch in `Verify`, single writer version.**
`Verify` switches on `schema_version`: `"3.0"` uses the full-document rule; `"2.0"` uses the legacy entries-only rule and returns a distinct warning the CLI must surface; anything else fails with "unsupported schema version". `Build` always writes `"3.0"`. Alternative: refusing 2.0 files outright — rejected, users may hold old exports and losing verification for them is worse than a warned legacy check.

**4. `--pubkey` accepts a base64 string or a file path.**
If the value parses as base64 of exactly 32 bytes, use it directly; otherwise treat it as a path and read either a PEM (`ED25519 PUBLIC KEY` block, matching `identity.pub`) or raw base64 content. Comparison is constant-time (`crypto/subtle`) against the document's embedded key, then normal signature verification proceeds. Without `--pubkey`, behavior is unchanged except the output labels the check as "consistency only (embedded key)".

**5. Exit codes and output.**
`verify` keeps exit 0 on success, non-zero on any failure including pubkey mismatch. Output lines state: schema version, signature scope (full document vs legacy entries-only), origin check (matched `--pubkey` / not requested).

## Risks / Trade-offs

- [Breaking change for external consumers pinned to `"2.0"`] → version bump is the documented signal; changelog entry; 2.0 files still verify.
- [Float canonicalization differs across future Go versions] → unit test pins the canonical bytes of a fixture document; a Go upgrade that changes float formatting fails loudly in CI.
- [Legacy 2.0 path keeps weak verification alive indefinitely] → acceptable; the warning makes the guarantee explicit, and no new 2.0 files are produced.

## Migration Plan

1. Implement `Build`/`Verify` changes behind the version constant; update schema JSON.
2. Update CLI flag and messages; add tests including a committed 2.0 fixture.
3. Update CLAUDE.md design decisions and the export spec via delta.
Rollback: revert the commit; 3.0 files produced in the interim will fail verification on the reverted binary (unsupported version) — acceptable for a local tool.

## Open Questions

None.
