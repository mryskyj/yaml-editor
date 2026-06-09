package frontend

import (
	"io/fs"
	"testing"
)

func TestAssetsIncludesIndexHTML(t *testing.T) {
	t.Parallel()

	if _, err := fs.Stat(Assets(), "index.html"); err != nil {
		t.Fatalf("Assets() missing index.html: %v", err)
	}
}

func TestAssetsIncludesMonacoLoader(t *testing.T) {
	t.Parallel()

	if _, err := fs.Stat(Assets(), "vs/loader.js"); err != nil {
		t.Fatalf("Assets() missing Monaco loader: %v", err)
	}
}
