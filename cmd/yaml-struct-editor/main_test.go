package main

import "testing"

func TestParseSchemaOptionsFlags(t *testing.T) {
	t.Parallel()

	options := parseSchemaOptions([]string{
		"--schema-dir", "/tmp/schema",
		"--schema-type", "CustomConfig",
	})
	if options.Dir != "/tmp/schema" {
		t.Fatalf("Dir = %q, want /tmp/schema", options.Dir)
	}
	if options.Type != "CustomConfig" {
		t.Fatalf("Type = %q, want CustomConfig", options.Type)
	}
}

func TestParseSchemaOptionsPositionalDir(t *testing.T) {
	t.Parallel()

	options := parseSchemaOptions([]string{"/tmp/schema"})
	if options.Dir != "/tmp/schema" {
		t.Fatalf("Dir = %q, want /tmp/schema", options.Dir)
	}
	if options.Type != "" {
		t.Fatalf("Type = %q, want empty", options.Type)
	}
}
