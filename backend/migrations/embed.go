// Package migrations expone los archivos SQL de migración embebidos en el binario.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
