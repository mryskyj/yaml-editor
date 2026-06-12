package main

import (
	"os"
	"reflect"
	"testing"
)

func TestParseStartupOptions(t *testing.T) {
	t.Parallel()

	options, err := parseStartupOptions([]string{
		"--schema-dir",
		`C:\schemas\sample`,
		"--schema-type",
		"Config",
	})
	if err != nil {
		t.Fatalf("parseStartupOptions() returned error: %v", err)
	}
	if options.schemaDir != `C:\schemas\sample` {
		t.Fatalf("schemaDir = %q", options.schemaDir)
	}
	if options.schemaType != "Config" {
		t.Fatalf("schemaType = %q", options.schemaType)
	}
}

func TestParseStartupOptionsRejectsUnexpectedArgument(t *testing.T) {
	t.Parallel()

	if _, err := parseStartupOptions([]string{"config.yaml"}); err == nil {
		t.Fatal("parseStartupOptions() error = nil, want error")
	}
}

func TestSanitizeProcessArgsForWails(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})

	os.Args = []string{"yaml-struct-editor.exe", "--schema-dir", `C:\schemas\sample`}
	sanitizeProcessArgsForWails()

	if want := []string{"yaml-struct-editor.exe"}; !reflect.DeepEqual(os.Args, want) {
		t.Fatalf("os.Args = %#v, want %#v", os.Args, want)
	}
}
