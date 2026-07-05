<!-- tokenpile-skill-version: 3 -->

# tokenpile

tokenpile tracks LLM token usage and cost per GitHub issue. You have two responsibilities: log your own usage, and answer questions about usage data by running CLI commands.

## 1. Logging usage

After each response where substantial work was done, log usage:

```
tokenpile log \
  --issue <issue-number> \
  --agent claude-code \
  --model <model-id> \
  --tokens-in <input-tokens> \
  --tokens-out <output-tokens> \
  --note "<one-line summary of what was done>" \
  --tag <tag> \
  [--repo owner/repo]
```

**Parameters:**
- `--issue`: GitHub issue number for the current task. Ask the user if unknown.
- `--agent`: always `claude-code`
- `--model`: current model, e.g. `claude-sonnet-4-6`, `claude-opus-4-8`, `claude-haiku-4-5`
- `--tokens-in` / `--tokens-out`: estimate based on this response. Count approximately 4 characters per token. For `--tokens-in`, estimate the total context fed into this turn (conversation history + files read + tool outputs). For `--tokens-out`, estimate the tokens in your response. Log automatically without asking the user — approximate counts are acceptable.
- `--note`: one-line description of what was done in this response (max 100 chars). Example: `"refactored auth middleware"`, `"fixed unicode bug in lexer"`. Always include.
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

**Example questions and how to handle them:**

- "How many tokens did I spend on issue #42?" → run `tokenpile report --issue 42`
- "Show me the sessions for this issue" → run `tokenpile report --issue <N> --sessions`
- "What did this session cost?" → run `tokenpile report --issue <current-issue>`
- "Show me usage for the last week" → run `tokenpile export --from <date>` or open the TUI with `tokenpile`
- "Am I over budget?" → run `tokenpile report --issue <N>` and check the budget line
- "Am I logged in?" → run `tokenpile auth status`

Always run the command and include the output in your response. Do not guess or estimate when real data is available.
