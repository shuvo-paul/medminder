---
name: create-issue-from-srs
description: Use when creating a GitHub issue from a specific SRS requirement ID (e.g. REQ-AUTH-001) or section reference (e.g. 3.1.1 or "Registration") in the MedMinder project.
---

# Create Issue from SRS

Create one or more GitHub issues from a specific requirement or section in `docs/SRS.md`.

## Input

A single SRS reference — one of:

- **Requirement ID**: `REQ-AUTH-001`
- **Section number**: `3.1.1`
- **Section name**: `Registration`

## Scope Detection

After locating the reference in `docs/SRS.md`, determine output type:

| Reference resolves to | Output |
|---|---|
| A single requirement (leaf REQ-ID) | One issue |
| A section containing multiple requirements | One parent issue + sub-issues (one per logical task group) |

A **logical task group** is a cluster of closely related requirements that would naturally be implemented together in the same PR (e.g., "create user + validate email" rather than one issue per line). Example: section `3.1.1 Registration` with REQ-AUTH-001 through REQ-AUTH-005 might group into two sub-issues: `[3.1.1.1] User registration flow` (REQ-AUTH-001–003) and `[3.1.1.2] Email verification` (REQ-AUTH-004–005).

## Process

1. **Read** `docs/SRS.md` — locate the referenced REQ-ID or section and extract all relevant requirements, their IDs, and descriptions
2. **Scan codebase** — check `internal/features/` and related files for anything already built that provides context
3. **Search existing issues** — run `gh issue list --search "<REQ-ID or section title>" --state all` to avoid creating duplicates
4. **Determine scope** — leaf requirement → single issue; section → parent + sub-issues
5. **Create issue(s)** — use templates below; if parent + sub-issues, create sub-issues first, then reference them in the parent body

## Single Issue Template

Title: `[REQ-XXX-NNN] <brief description from SRS>`

Body:

```
## Goal
[One sentence: what this requirement achieves for the user]

## Context
[What is already built in the codebase that is relevant to this work. If nothing is built yet, state that explicitly.]

## Requirements
- REQ-XXX-NNN: [full requirement text from SRS]
[add more REQ lines if this issue covers a tightly coupled pair]

## Tasks
- [ ] [implementation task derived from requirement]
- [ ] [...]

## Acceptance Criteria
- [ ] [verifiable condition the implementation must meet]
- [ ] [...]

## Out of Scope
- [explicitly excluded items — things a reader might expect but that are NOT included]

## Dependencies
- Blocked by: [#N or "none"]
- Blocks: [#N or "none"]
```

Labels: `feature`, `enhancement`

## Parent Issue Template (for sections)

Title: `[<Section number>] <Section name>`

Body:

```
## Goal
[What this section achieves when all sub-issues are complete]

## Sub-issues
- [ ] #N — [sub-issue title]
- [ ] #N — [sub-issue title]
[...]

## Requirements Covered
[Comma-separated list of all REQ-IDs in this section]

## Completion Criteria
All sub-issues closed and merged.
```

Labels: `feature`, `tracking`

## Sub-Issue Template

Title: `[<Section number>.<group index>] <task group name>`

Body: Use the **Single Issue Template** above. The Requirements section lists all REQ-IDs belonging to this task group.

Labels: `feature`, `enhancement`

## gh CLI Commands

Use a heredoc to pass multi-line issue bodies — inline `--body "..."` breaks on newlines and quotes:

```bash
# Create a single issue or sub-issue
gh issue create \
  --title "[REQ-AUTH-001] User registration" \
  --body "$(cat <<'EOF'
## Goal
...

## Context
...

## Requirements
- REQ-AUTH-001: ...

## Tasks
- [ ] ...

## Acceptance Criteria
- [ ] ...

## Out of Scope
- ...

## Dependencies
- Blocked by: none
- Blocks: none
EOF
)" \
  --label "feature,enhancement"

# Create a parent tracking issue
gh issue create \
  --title "[3.1.1] Registration" \
  --body "$(cat <<'EOF'
## Goal
...

## Sub-issues
- [ ] #12 — [3.1.1.1] Registration flow
- [ ] #13 — [3.1.1.2] Email verification

## Requirements Covered
REQ-AUTH-001, REQ-AUTH-002, REQ-AUTH-003, REQ-AUTH-004, REQ-AUTH-005

## Completion Criteria
All sub-issues closed and merged.
EOF
)" \
  --label "feature,tracking"

# Check for duplicates before creating
gh issue list --search "REQ-AUTH-001" --state all
```

> **Note:** If labels `feature`, `enhancement`, or `tracking` don't exist in the repo, create them first:
>
> ```bash
> gh label create feature --color 0075ca
> gh label create enhancement --color a2eeef
> gh label create tracking --color e4e669
> ```

## Quality Checks

Before creating any issue, verify:

- [ ] REQ-IDs in the issue body match exactly what is in `docs/SRS.md`
- [ ] "Context" section reflects actual current codebase state (not guessed)
- [ ] No duplicate issue exists for this REQ-ID
- [ ] Tasks checklist is derived from the requirements — not generic filler
- [ ] "Out of Scope" explicitly names at least one boundary
