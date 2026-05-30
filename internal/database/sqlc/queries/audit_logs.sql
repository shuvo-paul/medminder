-- name: CreateAuditLog :exec
INSERT INTO audit_logs (id, event_type, user_id, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6);
