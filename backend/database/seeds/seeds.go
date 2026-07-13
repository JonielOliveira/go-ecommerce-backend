// Package seeds embute os scripts SQL de seed de desenvolvimento no
// binário, para que possam ser executados sem depender do diretório de
// trabalho atual (ver cmd/seed).
package seeds

import _ "embed"

//go:embed seed.sql
var SeedSQL string
