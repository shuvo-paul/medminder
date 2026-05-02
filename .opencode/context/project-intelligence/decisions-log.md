<!-- Context: project-intelligence/decisions | Priority: high | Version: 1.0 | Updated: 2025-01-12 -->

# Decisions Log

> Record major architectural and business decisions with full context. This prevents "why was this done?" debates.

## Quick Reference

- **Purpose**: Document decisions so future team members understand context
- **Format**: Each decision as a separate entry
- **Status**: Decided | Pending | Under Review | Deprecated

## Decision Template

```markdown
## [Decision Title]

**Date**: YYYY-MM-DD
**Status**: [Decided/Pending/Under Review/Deprecated]
**Owner**: [Who owns this decision]

### Context
[What situation prompted this decision? What was the problem or opportunity?]

### Decision
[What was decided? Be specific about the choice made.]

### Rationale
[Why this decision? What were the alternatives and why were they rejected?]

### Alternatives Considered
| Alternative | Pros | Cons | Why Rejected? |
|-------------|------|------|---------------|
| [Alt 1] | [Pros] | [Cons] | [Why not chosen] |
| [Alt 2] | [Pros] | [Cons] | [Why not chosen] |

### Impact
**Positive**: [What this enables or improves]
**Negative**: [What trade-offs or limitations this creates]
**Risk**: [What could go wrong]

### Related
- [Links to related decisions, PRs, issues, or documentation]
```

---

## Decision: [Title]

**Date**: YYYY-MM-DD
**Status**: [Status]
**Owner**: [Owner]

### Context
[What was happening? Why did we need to decide?]

### Decision
[What we decided]

### Rationale
[Why this was the right choice]

### Alternatives Considered
| Alternative | Pros | Cons | Why Rejected? |
|-------------|------|------|---------------|
| [Option A] | [Good things] | [Bad things] | [Reason] |
| [Option B] | [Good things] | [Bad things] | [Reason] |

### Impact
- **Positive**: [What we gain]
- **Negative**: [What we trade off]
- **Risk**: [What to watch for]

### Related
- [Link to PR #000]
- [Link to issue #000]
- [Link to documentation]

---

## Decision: DTO Extraction from Handlers

**Date**: 2026-05-02
**Status**: Decided
**Owner**: MedMinder Core Team

### Context
The auth feature had handler types (`RegisterInput`, `LoginInput`, `LogoutInput`, etc.) living in `handlers/` packages, but they served double duty — they were both the Huma HTTP contract (with validation struct tags) AND the domain types consumed by handler functions. This caused several problems: (1) `routes.go` had to do sentinel→Huma error translation, bleeding HTTP concerns into the wiring layer; (2) output structs were duplicated inline across multiple routes; (3) the boundary between "wire format" and "handler domain" was blurred.

### Decision
Introduce a dedicated `dto` package per feature (`internal/features/<feature>/dto/`) that owns all request/response types with Huma struct tags. Handlers accept `*dto.Input` types and return `*dto.Output` types directly. Error translation (sentinel errors → `huma.Error*`) lives inside the handler, not in `routes.go`. Routes become pure wiring that passes handlers directly to `huma.Register`.

### Rationale
Separation: `dto` = wire format + validation, `handlers` = HTTP↔service adapter, `routes.go` = wiring only. Each package has one job. This follows the adapter pattern from Clean Architecture — handlers are the primary adapter between HTTP and the service layer. Routes were doing too much.

### Alternatives Considered
| Alternative | Pros | Cons | Why Rejected? |
|-------------|------|------|---------------|
| Keep handler types as DTOs | Less files, simple structure | Blurs boundaries, error translation in routes | Violates single responsibility |
| DTOs in routes.go | Keeps handlers thin | Still no shared types, routes gets complex | Routes is not the right place for domain types |
| Flat handler package with all types | Single package to import | Giants package, poor locality | Goes against Go package design principles |

### Impact
- **Positive**: Clean boundaries, handlers own HTTP semantics, routes are 1-liners, error translation at the right layer
- **Negative**: Extra `dto/` package per feature (minimal)
- **Risk**: None identified — the pattern is proven in the codebase already (`password_reset.go` already used DTOs)

### Related
- PR: feature/issue-60-auth-service-refactor (`8eef3a8`)
- `technical-domain.md` — updated to reflect `dto/` package

---

## Decision: [Title]

**Date**: YYYY-MM-DD
**Status**: [Status]
**Owner**: [Owner]

### Context
[What was happening?]

### Decision
[What we decided]

### Rationale
[Why this was right]

### Alternatives Considered
| Alternative | Pros | Cons | Why Rejected? |
|-------------|------|------|---------------|
| [Option A] | [Good things] | [Bad things] | [Reason] |

### Impact
- **Positive**: [What we gain]
- **Negative**: [What we trade off]

### Related
- [Link]

---

## Deprecated Decisions

Decisions that were later overturned (for historical context):

| Decision | Date | Replaced By | Why |
|----------|------|-------------|-----|
| [Old decision] | [Date] | [New decision] | [Reason] |

## Onboarding Checklist

- [ ] Understand the philosophy behind major architectural choices
- [ ] Know why certain technologies were chosen over alternatives
- [ ] Understand trade-offs that were made
- [ ] Know where to find decision context when questions arise
- [ ] Understand what decisions are pending and why

## Related Files

- `technical-domain.md` - Technical implementation affected by these decisions
- `business-tech-bridge.md` - How decisions connect business and technical
- `living-notes.md` - Current open questions that may become decisions
