---
name: tokenpile
description: Track LLM token usage and cost per GitHub issue. Use after any response where you did meaningful work to log token usage, and whenever the user asks about token usage, cost, sessions, or spending budget for an issue.
---

<!-- tokenpile-skill-version: 4 -->

# tokenpile

tokenpile tracks LLM token usage and cost per GitHub issue. Log your own usage after each response, and answer usage questions by running CLI commands.

## 1. Logging usage

After each response where substantial work was done:

```
tokenpile log \
  --issue <issue-number> \
  --agent codex \
  --model <model-id> \
  --tokens-in <input-tokens> \
  --tokens-out <output-tokens> \
  --note "<one-line summary of what was done>" \
  --tag <tag> \
  [--repo owner/repo]
```

**Parameters:**
- `--issue`: GitHub issue number for the current task. Ask the user if unknown.
- `--agent`: always `codex`
- `--model`: current model, e.g. `o3`, `o4-mini`, `gpt-4o`
- `--tokens-in` / `--tokens-out`: estimate (~4 chars per token). For `--tokens-in`, estimate total context for this turn (conversation history + files read + tool outputs). For `--tokens-out`, estimate tokens in your response. Approximate counts are acceptable. Log automatically without asking the user.
- `--note`: one-line description of what was done in this response (max 100 chars). Always include.
- `--tag`: one or more tags from this vocabulary (repeat the flag for multiple): `refactor`, `debug`, `feature`, `test`, `docs`, `spike`, `review`. Choose all that apply.
- `--repo`: optional if running inside a git repo with a GitHub remote

**When to log:**
- At the end of a response where you used tools, wrote code, or did meaningful analysis
- Once per user turn, not after every tool call
- Skip for one-liner answers or trivial replies

**Sessions are automatic:** consecutive logs within 30 minutes of the previous log share the same session. Tags accumulate across log calls in the same session (union); the note is overwritten by the latest call. No action needed from you.

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

Always run the command and include the output in your response. Do not guess or estimate when real data is available.
