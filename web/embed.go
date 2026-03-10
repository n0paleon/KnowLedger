package web

import "embed"

var (
	//go:embed views static
	FS embed.FS
)
