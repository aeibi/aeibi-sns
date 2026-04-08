package web

import "embed"

// DistFS contains the built frontend assets.
//
//go:embed dist
var DistFS embed.FS
