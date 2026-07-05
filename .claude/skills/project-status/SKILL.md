---
name: project-status
description: On-demand project status report covering recent work, current state, and drift between code, docs, and specs.
license: MIT
compatibility:
  - claude-code
metadata:
  author: cdimonaco
  version: 1
  generatedBy: human
---

# Project Status Report

Generate a concise status report for the tokenpile project. Covers what was done recently, what is active now, and whether code, docs, and specs appear to be in sync.

## Steps

1. Run these commands in parallel:

   ```bash
   git log --oneline -15
   git status --short
   git branch --show-current
   openspec list --json 2>/dev/null || echo '{"changes":[]}'
   gh run list --limit 3 --json status,name,conclusion,createdAt 2>/dev/null || echo '[]'
   ```

2. Read in parallel:
   - `README.md` — features list and commands section (first 120 lines is enough)
   - `openspec/specs/` — list the directory to see which capabilities are documented

3. Analyze for drift. Check each of these:
   - Do recent commits (feat: or fix:) touch packages not mentioned in README features?
   - Are there active OpenSpec changes with no corresponding recent commits?
   - Are there OpenSpec specs for capabilities that appear to be unimplemented (no matching package in `internal/`)?

4. Produce the report in this exact format:

---
## Project Status — tokenpile

### Recent work
<list the last 15 commits grouped by type: feat, fix, test, chore, docs, refactor>

### Current state
- Branch: <current branch>
- Uncommitted: <"clean" or list of files>
- CI: <last run name + conclusion, or "unknown">
- Active OpenSpec changes: <list change names or "none">

### Documented capabilities
<one line per file found in openspec/specs/, e.g. "usage-tracking — core entry/session model">

### Drift signals
<bullet list of concrete inconsistencies found, or "none detected">
---

## Rules

- Flag drift only when there is clear evidence. Do not speculate.
- Keep each section tight — one line per item where possible.
- If a tool is unavailable (openspec, gh), note it in the relevant section and continue.
- Do not implement or change anything. This skill is read-only.
