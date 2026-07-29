## 1. Remove the agent

- [x] 1.1 Delete the `codex` entry from the `agents` slice in `internal/skill/skill.go`, including its `InstallPath` and `LegacySharedPath`
- [x] 1.2 Remove the `//go:embed templates/codex.md` directive and the `codexTemplate` variable
- [x] 1.3 Delete `internal/skill/templates/codex.md`
- [x] 1.4 Confirm no other reference to codex remains in Go code (`grep -rn codex --include='*.go'`)

## 2. Tests

- [x] 2.1 Remove codex cases from the skill tests
- [x] 2.2 Add a test asserting `Install("codex")` returns `ErrUnsupportedAgent` and writes no file, so the removal is pinned rather than merely absent
- [x] 2.3 Add a test asserting `List()` yields exactly `claude-code` and `opencode`
- [x] 2.4 Verify the reset integration tests still pass unchanged: `reset` iterates `skill.List()` and needs no edit

## 3. Docs and specs

- [x] 3.1 README: supported agents list drops codex
- [x] 3.2 When syncing to the main spec, update the `agent-skill` Purpose line, which still names codex alongside Claude Code and opencode
- [x] 3.3 Run `make check` and commit
