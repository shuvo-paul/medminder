<!-- Context: project-intelligence/nav | Priority: critical | Version: 2.0 | Updated: 2026-05-01 -->

# Project Intelligence

> Start here for quick project understanding. These files bridge business and technical domains for MedMinder.

## Structure

```
.opencode/context/project-intelligence/
├── navigation.md              # This file — quick overview
├── business-domain.md         # Business context and problem statement
├── technical-domain.md        # Stack, architecture, patterns, standards ⭐
├── business-tech-bridge.md    # How business needs map to solutions
├── decisions-log.md           # Major decisions with rationale
└── living-notes.md            # Active issues, debt, open questions
```

## Quick Routes

| What You Need | File | Priority |
|---------------|------|----------|
| Understand the "why" | `business-domain.md` | high |
| Understand the "how" | `technical-domain.md` | **critical** |
| See the connection | `business-tech-bridge.md` | high |
| Know the context | `decisions-log.md` | medium |
| Current state | `living-notes.md` | medium |

## Deep Dives

| Topic | File | Description |
|-------|------|-------------|
| Tech Stack & Versions | `technical-domain.md` → Primary Stack | Go 1.25 + Chi + huma v2 + PostgreSQL + SvelteKit |
| API Patterns | `technical-domain.md` → Code Patterns | huma v2 typed handlers `func(ctx, *Input) (*Output, error)` |
| Component Patterns | `technical-domain.md` → Code Patterns | SvelteKit pages with shadcn-svelte, $state() runes |
| Naming Conventions | `technical-domain.md` → Naming Conventions | snake_case files, PascalCase exports, kebab-case routes |
| Standards | `technical-domain.md` → Code Standards | TDD, feature-module layout, Makefile, git flow |
| Security | `technical-domain.md` → Security Requirements | JWT auth, input validation, parameterized queries |

## Usage

**New Developer / Agent**:
1. Start with `navigation.md` (this file)
2. Read `technical-domain.md` thoroughly — it has the full tech map
3. Follow codebase references to see real implementations

**Quick Handoff**:
- Building a feature → read `technical-domain.md` → see patterns → read feature examples
- Debugging → `decisions-log.md` + `living-notes.md`

## Integration

Referenced from:
- `.opencode/context/core/standards/project-intelligence.md` — standards and patterns
- `.opencode/context/core/system/context-guide.md` — context loading

## Maintenance

- **Update when**: Tech stack changes, new features added, patterns evolve
- **Version tracking**: 2.0 for structure change (replaced template with real data)
- **Review trigger**: `/add-context --update` or significant architecture changes
