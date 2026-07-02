# tokenpile usage tracking

After each response where you make significant LLM calls, log usage with the tokenpile CLI:

```
tokenpile log \
  --issue <issue-number> \
  --agent claude-code \
  --model <current-model-id> \
  --tokens-in <input-token-count> \
  --tokens-out <output-token-count> \
  [--repo owner/repo]
```

## Notes

- `--issue` is the GitHub issue number you are working on (required, ask the user if unknown)
- `--agent` must be `claude-code` (fixed for this skill)
- `--model` is the model identifier, e.g. `claude-sonnet-4-6`, `claude-opus-4-7`
- `--tokens-in` and `--tokens-out` are the token counts for the current session/response
- `--repo` is optional if you are inside a git repository with a GitHub remote; otherwise pass it explicitly

## When to log

Log once per substantial interaction, not after every tool call. A good trigger is at the end of a user-facing response where meaningful work was done.
