// Package migrations provides embedded database migrations.
package migrations

import (
	"embed"
)

// FS embeds the SQL migration files.
//
//go:embed *.sql
var FS embed.FS
