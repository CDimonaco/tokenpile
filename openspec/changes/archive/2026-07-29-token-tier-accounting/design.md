## Context

`usage.Entry` has `TokensIn int` and `TokensOut int`. `usage_entries` declares both `NOT NULL`. `pricing.ModelPrice` has `InputPerMillion` and `OutputPerMillion`, and `ComputeCost(model string, tokensIn, tokensOut int) (float64, bool)` multiplies one by the other. Cost is computed at report time and never stored — that decision stands and this change depends on it, since recomputing history under new rates is what makes a schema change survivable at all.

The production dataset is one entry, one session, zero budgets. There is nothing to migrate and no user to keep compatible, which removes every constraint that would normally shape a change of this size.

## Goals / Non-Goals

**Goals:**
- Represent what providers actually bill, so cost is correct rather than approximately wrong.
- Make cache savings and per-tier cost share computable, since those are the questions the tool exists to answer.
- Record whether a number was measured or estimated, so a signed export attests to provenance and not only authorship.
- Design the schema as if writing it today: no legacy fields, no compatibility shims.

**Non-Goals:**
- Capturing the tiers. Nothing in this change fills them from an agent transcript; `log` remains the only writer and remains `estimated`. Capture is the next change, and this one exists to give it somewhere to write.
- Making `issue_num` nullable. That belongs with capture.
- Per-provider pricing inheritance or a provider abstraction in the pricing config.
- Migrating existing rows.

## Decisions

**1. `Usage` as a value type, not five fields on `Entry`.**
```go
// Usage is the token accounting for a single entry, as reported by the
// provider. Reasoning is a SUBSET of Output, never added on top.
type Usage struct {
    InputFresh int
    CacheWrite int
    CacheRead  int
    Output     int
    Reasoning  int
}
```
It travels intact through `Entry`, `ReportRow`, `Point` and `TrackedIssue`, and gives cost computation one parameter instead of five. Five loose ints on five types is the shape that makes a sixth tier a five-file change.

**2. `Reasoning` is a subset of `Output`.**
OpenAI reports reasoning tokens inside `completion_tokens`; adding them to output double-bills. This is the single most likely defect in the change, it is invisible in every test that only checks totals, and it is worth an explicit test that a `Usage` with `Output: 100, Reasoning: 40` costs the same as one with `Output: 100, Reasoning: 0`.
Alternative considered: omit `Reasoning` entirely. Rejected — opencode already exposes it per message, and adding it later means recreating the table a second time, at which point real data exists.

**3. Explicit cache rates, not multipliers.**
```yaml
claude-opus-5:
  input_per_million:       5.00
  output_per_million:      25.00
  cache_read_per_million:  0.50
  cache_write_per_million: 6.25
```
Multipliers (`cache_read_multiplier: 0.10`) would be more compact and are stable across Anthropic's line, but they assume cache pricing is always a ratio of the input rate. OpenAI's ratios differ, and nothing stops a provider from pricing cache absolutely. Explicit figures also mean every number can be checked against a published price list without arithmetic. The cost is 96 numbers in the defaults instead of 48.

**4. Missing cache rates warn; they do not default.**
When an entry carries cache tokens for a model with no cache rates, treating them as input overstates by roughly 7x and assuming 0.10 silently encodes an assumption that the provider is Anthropic. Neither is acceptable as a silent default, so this reuses the existing unknown-model warning: WARN via `slog` plus a visible marker in report output. Being told the number is incomplete beats being given a confident wrong one.

**5. `source` is derived from the writer, not passed as a flag.**
`log` writes `estimated`; the capture path introduced by the next change writes `measured`. A `--source` flag would let a caller assert something it cannot know and would eventually be set wrong by a script. Deriving it means the field says how the row actually arrived.
This is what turns the Ed25519 signature into a claim worth making: today it attests who produced a number, not whether the number was observed or invented.

**6. `ComputeCost` takes the struct.**
```go
func (l *Loader) ComputeCost(model string, u usage.Usage) (float64, bool)
```
Four tiers make positional parameters unreadable and trivially transposable at a call site. Mechanical change, touches every caller.

**7. Report defaults to total cost and cache savings; tiers behind `--detail`.**
Cache savings — what the same tokens would have cost at the fresh input rate, minus what they cost — is the line worth seeing every time. The four-tier table is diagnostics. Putting both on screen by default makes the common case noisy for the sake of the rare one.

**8. Export goes to 4.0 and drops verification of every earlier version.**
Entry objects gain per-tier fields and `source`; `tokens_in`/`tokens_out` disappear rather than being retained as derived sums, since keeping them would invite consumers to depend on a number with no owner.

Keeping 2.0 and 3.0 verification looked free and is not. `documentDigest` canonicalizes the *parsed* `Document`, not the bytes on disk, so the digest depends on the fields the Go types define today. A 3.0 entry parsed by the new types loses `tokens_in`/`tokens_out` and gains five tier fields as zeros, which changes the canonical form and invalidates the signature. Verified empirically before deciding.

Two ways out were available: verify from raw bytes, which would make verification independent of the Go types for good, or keep a legacy entry type per schema version. Both were rejected in favour of dropping pre-4.0 verification outright, on the owner's decision that signature compatibility with old documents is not wanted. The raw-bytes approach remains the right fix if that ever changes.

Worth recording: the 2.0 path survived earlier schema changes by accident, not design. Signing a Go value rather than a byte sequence means every schema change silently breaks verification of everything written before it.

**9. The store is recreated, not migrated.**
The `migrations` slice of additive `ALTER TABLE` statements stays for future additive changes, but this change edits `CREATE TABLE` directly. With one row in existence, a migration path would be untested code written to serve nobody.

## Risks / Trade-offs

- [Reasoning double-counted, inflating every cost silently] → pinned by an explicit equality test; called out here as the most probable defect.
- [96 pricing numbers to maintain, and wrong ones are invisible] → each is checkable against a published list, which multipliers are not; the unknown/missing-rate warning covers the gap until a model is filled in.
- [Cost is notional under subscription plans] → the figure is what the tokens would cost at list rates, which is not what a Max-plan user pays. This is already true today and this change does not fix it, but making the numbers credible makes the ambiguity more dangerous, so it is worth naming the field `cost` deliberately or renaming it in a later change.
- [Every token-touching test is rewritten at once] → unavoidable, and cheaper now than after the capture change lands and the suite is larger.
- [Nothing yet writes `measured`] → intentional; this change is the schema, the next one is the writer. The `source` field ships with a single possible value until then, which is honest rather than premature.
