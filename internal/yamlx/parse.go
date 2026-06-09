package yamlx

import (
	"fmt"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

var yamlLinePattern = regexp.MustCompile(`line ([0-9]+)`)

// Diagnostic describes a YAML parser issue in editor-friendly coordinates.
type Diagnostic struct {
	Message string
	Line    int
	Column  int
}

// Document wraps a parsed YAML document node.
type Document struct {
	Root *yaml.Node
}

// Content returns the root content node inside the YAML document node.
func (d *Document) Content() *yaml.Node {
	if d == nil || d.Root == nil || len(d.Root.Content) == 0 {
		return nil
	}

	return d.Root.Content[0]
}

// Parse parses YAML source and returns syntax or unsupported-feature diagnostics.
func Parse(source string) (*Document, []Diagnostic) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(source), &root); err != nil {
		return nil, []Diagnostic{diagnosticFromError(err)}
	}

	document := &Document{Root: &root}
	return document, UnsupportedDiagnostics(&root)
}

// UnsupportedDiagnostics reports unsupported YAML features in a parsed node tree.
func UnsupportedDiagnostics(root *yaml.Node) []Diagnostic {
	var diagnostics []Diagnostic
	walk(root, func(node *yaml.Node) {
		if node.Kind == yaml.AliasNode {
			position := NodePosition(node)
			diagnostics = append(diagnostics, Diagnostic{
				Message: "YAML Alias is not supported",
				Line:    position.Line,
				Column:  position.Column,
			})
		}
	})

	return diagnostics
}

func walk(node *yaml.Node, visit func(*yaml.Node)) {
	if node == nil {
		return
	}

	visit(node)
	for _, child := range node.Content {
		walk(child, visit)
	}
}

func diagnosticFromError(err error) Diagnostic {
	line := 1
	matches := yamlLinePattern.FindStringSubmatch(err.Error())
	if len(matches) == 2 {
		if parsedLine, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			line = parsedLine
		}
	}

	return Diagnostic{
		Message: fmt.Sprintf("YAML syntax error: %s", err.Error()),
		Line:    line,
		Column:  1,
	}
}
