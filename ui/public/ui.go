package public

import "embed"

// Content holds our static web server content.
//go:embed *
var Content embed.FS
