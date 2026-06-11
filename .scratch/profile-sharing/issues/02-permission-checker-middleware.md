Status: done

## What to build

Implement the permission checker service and middleware that enforces permission-based access control on profile-scoped endpoints.

### Permission checker service
A function that determines whether a user has a specific permission on a profile. Algorithm:
1. Query `profile_permissions` for `(profile_id, shared_with_user_id)`
2. If no row exists → return false
3. If `status != 'accepted'` → return false
4. If `expires_at` is set and in the past → return false
5. Check if `permissions` JSONB array contains the required permission(s)

### Permission middleware
- Extract user ID from JWT (already available via existing auth middleware)
- Extract profile ID from URL path (e.g., `/api/profiles/{id}/*`)
- Query permissions table
- If user has the required permission, attach permission context to request and proceed
- If not, return 403 Forbidden

The middleware should be configurable with which permission(s) are required per route. Support both single permission checks and "any of these permissions" checks (e.g., `profile:share` OR `profile:admin`).

Support edge cases:
- Guest access path is separate — guest tokens don't go through this middleware
- Listing endpoints (e.g., `GET /api/profiles`) aren't middleware-checked — they use service-level filtering
- The middleware applies to all `/api/profiles/{id}/*` paths where authorization is needed

## Acceptance criteria

- [x] Permission checker returns correct results for: owner (all permissions), shared user (specific permissions), no access (false), expired (false), pending (false)
- [x] Middleware correctly handles single and "any-of" permission requirements
- [x] Middleware extracts user ID from JWT and profile ID from path
- [x] Forbidden responses return 403 with descriptive error
- [x] Listing endpoints are NOT blocked by middleware (handled at service layer)
- [x] Unit tests for permission checker, integration tests for middleware

## Blocked by

- #1 Permission DB Schema
