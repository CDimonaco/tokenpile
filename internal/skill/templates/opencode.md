# tokenpile

tokenpile tracks LLM token usage and cost per GitHub issue. Log your own usage after each response, and answer usage questions by running CLI commands.

## 1. Logging usage

After each response where substantial work was done:

```
tokenpile log \
  --issue <issue-number> \
  --agent opencode \
  --model <model-id> \
  --tokens-in <input-tokens> \
  --tokens-out <output-tokens> \
  [--repo owner/repo]
```

**Parameters:**
- `--issue`: GitHub issue number for the current task. Ask the user if unknown.
- `--agent`: always `opencode`
- `--model`: the model currently in use (e.g. `claude-sonnet-4-6`, `gpt-4o`, `o3`). Use the exact model identifier configured in opencode.
- `--tokens-in` / `--tokens-out`: estimate (~4 chars per token). For `--tokens-in`, estimate total context for this turn (conversation history + files read + tool outputs). For `--tokens-out`, estimate tokens in your response. Approximate counts are acceptable. Log automatically without asking the user.
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

**Export data:**
```
tokenpile export [--issue <N>] [--repo owner/repo] [--from <RFC3339>] [--to <RFC3339>]
```

**Check auth status:**
```
tokenpile auth status
```

Always run the command and include the output in your response. Do not guess or estimate when real data is available.
