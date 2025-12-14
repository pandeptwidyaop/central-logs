package web

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var distFS embed.FS

// GetFileSystem returns the embedded frontend files
func GetFileSystem() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
