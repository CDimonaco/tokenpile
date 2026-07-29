---
name: tokenpile
description: Track LLM token usage and cost per GitHub issue. Use to declare which GitHub issue the current work belongs to, and whenever the user asks about token usage, cost, sessions, or spending budget for an issue.
---

<!-- tokenpile-skill-version: 6 -->

# tokenpile

tokenpile tracks LLM token usage and cost per GitHub issue. You have two responsibilities: declare which issue the work belongs to, and answer questions about usage data by running CLI commands.

**You do not report token counts.** A capture hook reads them from this session's transcript, where the provider recorded what it actually billed. You cannot see your own context window, tool definitions or prompt cache, and the cache alone routinely accounts for the large majority of tokens billed, so any figure you produced would be wrong by a wide margin.

## 1. Declaring the issue

As soon as you know which GitHub issue the work belongs to, declare it once:

```
tokenpile bind \
  --issue <issue-number> \
  --note "<one-line summary of what is being worked on>" \
  --tag <tag> \
  [--repo owner/repo]
```

**Parameters:**
- `--issue`: GitHub issue number for the current task. Ask the user if unknown.
- `--note`: one-line description of the work (max 100 chars). Example: `"refactored auth middleware"`. Always include.
- `--tag`: one or more tags from this vocabulary (repeat the flag for multiple): `refactor`, `debug`, `feature`, `test`, `docs`, `spike`, `review`. Choose all that apply.
- `--repo`: optional if running inside a git repo with a GitHub remote

**When to bind:**
- Once, as soon as the issue is known — not after every response
- Again only if the work moves to a different issue

**If you never bind, nothing is lost.** Usage is still captured with exact token counts; it is simply recorded as unattributed, and can be assigned later with `tokenpile unattributed`. Forgetting costs attribution, never measurement.

## 2. Answering questions about usage

When the user asks about token usage, cost, sessions, or budget, run the appropriate command and show the output.

**Report for a specific issue:**
```
tokenpile report --issue <N> [--repo owner/repo]
```
Shows per-agent, per-model breakdown with tokens, cost, and wall-clock time. If a budget is set, shows how much has been consumed.

**Per-session breakdown:**
```
tokenpile report --issue <N> --sessions [--repo owner/repo]
```
Shows each session with start/end time, duration, tags, and note.

**Manage spending budget:**
```
tokenpile budget set --issue <N> --amount <USD>
tokenpile budget unset --issue <N>
```

**Export data:**
```
tokenpile export [--issue <N>] [--repo owner/repo] [--from <RFC3339>] [--to <RFC3339>]
```

**Check auth status:**
```
tokenpile auth status
```

**Example questions and how to handle them:**

- "How many tokens did I spend on issue #42?" → run `tokenpile report --issue 42`
- "Show me the sessions for this issue" → run `tokenpile report --issue <N> --sessions`
- "What did this session cost?" → run `tokenpile report --issue <current-issue>`
- "Show me usage for the last week" → run `tokenpile export --from <date>` or open the TUI with `tokenpile`
- "Am I over budget?" → run `tokenpile report --issue <N>` and check the budget line
- "Am I logged in?" → run `tokenpile auth status`

Always run the command and include the output in your response. Do not guess or estimate when real data is available.
