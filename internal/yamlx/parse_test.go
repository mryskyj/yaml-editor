package yamlx

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseValidYAML(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse("server:\n  host: localhost\n  port: 8080\n")
	if len(diagnostics) != 0 {
		t.Fatalf("Parse() diagnostics = %#v, want none", diagnostics)
	}
	if document == nil {
		t.Fatal("Parse() document = nil")
	}
	if document.Root == nil || document.Root.Kind != yaml.DocumentNode {
		t.Fatalf("document.Root = %#v, want document node", document.Root)
	}

	content := document.Content()
	if content == nil || content.Kind != yaml.MappingNode {
		t.Fatalf("Content() = %#v, want mapping node", content)
	}
}

func TestParseSyntaxError(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse("server:\n  host: localhost\n  port\n    bad: value\n")
	if document != nil {
		t.Fatalf("Parse() document = %#v, want nil", document)
	}
	if len(diagnostics) != 1 {
		t.Fatalf("Parse() diagnostics count = %d, want 1", len(diagnostics))
	}

	diagnostic := diagnostics[0]
	if diagnostic.Line == 0 {
		t.Fatalf("diagnostic.Line = %d, want non-zero", diagnostic.Line)
	}
	if diagnostic.Column == 0 {
		t.Fatalf("diagnostic.Column = %d, want non-zero", diagnostic.Column)
	}
	if !strings.Contains(diagnostic.Message, "YAML syntax error") {
		t.Fatalf("diagnostic.Message = %q", diagnostic.Message)
	}
}

func TestParseAllowsAnchorAndDetectsAlias(t *testing.T) {
	t.Parallel()

	document, diagnostics := Parse("defaults: &defaults\n  host: localhost\nserver: *defaults\n")
	if document == nil {
		t.Fatal("Parse() document = nil")
	}
	if len(diagnostics) != 1 {
		t.Fatalf("Parse() diagnostics count = %d, want 1: %#v", len(diagnostics), diagnostics)
	}

	messages := map[string]bool{}
	for _, diagnostic := range diagnostics {
		if diagnostic.Line == 0 || diagnostic.Column == 0 {
			t.Fatalf("diagnostic position = %d:%d, want non-zero", diagnostic.Line, diagnostic.Column)
		}
		messages[diagnostic.Message] = true
	}

	if !messages["YAML Alias is not supported"] {
		t.Fatalf("missing alias diagnostic: %#v", diagnostics)
	}
}

func TestNodePosition(t *testing.T) {
	t.Parallel()

	node := &yaml.Node{Line: 3, Column: 7}
	position := NodePosition(node)
	if position.Line != 3 || position.Column != 7 {
		t.Fatalf("NodePosition() = %#v, want line 3 column 7", position)
	}
}

func TestNodePositionNil(t *testing.T) {
	t.Parallel()

	position := NodePosition(nil)
	if position.Line != 0 || position.Column != 0 {
		t.Fatalf("NodePosition(nil) = %#v, want zero position", position)
	}
}
