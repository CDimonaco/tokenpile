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
  [--repo owner/repo]
```

**Parameters:**
- `--issue`: GitHub issue number for the current task. Ask the user if unknown.
- `--agent`: always `claude-code`
- `--model`: current model, e.g. `claude-sonnet-4-6`, `claude-opus-4-8`, `claude-haiku-4-5`
- `--tokens-in` / `--tokens-out`: exact token counts for this response. Before logging, ask the user to run `/usage` in Claude Code and share the output. Parse the input/output token counts from that output. Do not estimate — always wait for the real numbers from `/usage`.
- `--repo`: optional if running inside a git repo with a GitHub remote

**When to log:**
- At the end of a response where you used tools, wrote code, or did meaningful analysis
- Once per user turn, not after every tool call
- Skip for one-liner answers or trivial replies

**Sessions are automatic:** consecutive logs within 30 minutes share the same session. No action needed.

## 2. Answering questions about usage

When the user asks about token usage, cost, or sessions, run the appropriate command and show the output.

**Report for a specific issue:**
```
tokenpile report --issue <N> [--repo owner/repo]
```
Shows per-agent, per-model breakdown with tokens, cost, and wall-clock time.

**List all tracked issues:**
```
tokenpile report --issue <N>
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
- "What did this session cost?" → run `tokenpile report --issue <current-issue>`
- "Show me usage for the last week" → run `tokenpile export --from <date>` or open TUI with `tokenpile`
- "Am I logged in?" → run `tokenpile auth status`

Always run the command and include the output in your response. Do not guess or estimate when real data is available.
