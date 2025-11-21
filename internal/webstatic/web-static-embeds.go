package webstatic

import (
	"embed"
)

//go:embed files/*
var files embed.FS

func FS() embed.FS {
	return files
}

func FSEmbedPath() string {
	return "files"
}
