## Context

`deterministic-capture` made `issue_num` nullable and added `ListUnattributed`, `AssignIssue` and `UnassignIssue`, grouped by session. `capture.Turn` already carries `Branch`, filled from Claude Code's `gitBranch` transcript field or resolved by the opencode plugin, but the reconciler uses it only to infer an issue and then drops it: `usage.Entry` has nowhere to put it.

That omission is why the reconciliation requirement in `deterministic-capture` was narrowed before archiving. Its spec now states plainly that suggestions are out of scope because the branch is not stored.

The TUI is Bubble Tea with a view enum, an issue list, a detail view and a chart view, each with its own key bindings.

## Goals / Non-Goals

**Goals:**
- Make reconciliation something a person will actually do, rather than something they could do.
- Deliver the suggestion the original design promised, from stored data rather than guesswork at display time.
- Keep the CLI path working unchanged; the TUI is an additional surface, not a replacement.

**Non-Goals:**
- Changing how capture works, what it reads, or the spool.
- Resolving a branch to a pull request to a linked issue over the API. Suggestions stay offline and pattern-based, as at capture time.
- Editing entries individually. The session stays the unit of assignment.
- Backfilling the branch for entries captured before this change: they keep an empty branch and simply offer no suggestion.

## Decisions

**1. Store the branch on the entry, do not re-derive it.**
The branch belongs to the moment of capture: it is what was checked out when the tokens were spent. Reading it back from the working directory at listing time would report today's branch, not that one, and would be wrong for exactly the sessions worth reconciling — old ones.
Alternative considered: store the branch on the session rather than the entry. Rejected — a session can outlive a branch switch, and the entry is where the other captured facts already live.

**2. The suggestion is derived on read, not stored.**
`InferFromBranch` is pure and cheap, and the mapping from branch names to issues is a heuristic that will be tuned. Storing its output would freeze a guess made under an older rule and require a migration to improve it; deriving it means every listing reflects the current rule.

**3. Group by session and branch, not by time window.**
The original wording said "repository, branch and time window". A session already is the time window: it opens on first use and idle-closes after thirty minutes. Adding a separate windowing rule would create groups that cut across sessions and break the promise that assigning a group moves its session too.

**4. The view surfaces itself.**
Unattributed usage counts toward no budget, so it is invisible spending until someone looks. `report` already says so in text; the issue list gains the same signal and a key to jump to the view. A reconciliation surface nobody navigates to is worth as little as no surface.

**5. Assignment in the view reuses the store methods verbatim.**
`AssignIssue` and `UnassignIssue` are already transactional and move the session with its entries. The view calls them; it does not grow its own path to the database. That also keeps the CLI and the TUI incapable of disagreeing.

**6. Entries captured before this change keep an empty branch.**
No backfill: the branch at capture time is unknowable after the fact, and inventing one would produce suggestions that look authoritative and are fiction. Those groups appear with no suggestion, which is honest.

## Risks / Trade-offs

- [Another schema change so soon after the last one] → the dataset is still tiny, and the alternative is a reconciliation surface that cannot do the one thing that makes it useful.
- [A wrong suggestion is worse than none, because it invites a careless confirmation] → the suggestion is always visible and editable before assigning, never applied automatically, and assignment is reversible.
- [The branch is a weak signal on repositories that do not encode issues in branch names] → those groups show no suggestion and are assigned manually, exactly as they are today.
- [TUI surface area grows] → the view is a list plus two actions, both delegating to store methods that already exist and are tested.
