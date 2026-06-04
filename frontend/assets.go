package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embeddedAssets embed.FS

// Assets returns the built frontend assets without the dist path prefix.
func Assets() fs.FS {
	assets, err := fs.Sub(embeddedAssets, "dist")
	if err != nil {
		panic(err)
	}
	return assets
}
