package web

import (
	"embed"
	"io/fs"
)

//go:embed ui
var embeddedUI embed.FS

func mustUIFS() fs.FS {
	sub, err := fs.Sub(embeddedUI, "ui")
	if err != nil {
		panic(err)
	}
	return sub
}
