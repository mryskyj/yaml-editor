package yamlx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const includeTag = "!include"

// ParseWithCommonInclude parses YAML and resolves common: !include "relative/path.yaml".
func ParseWithCommonInclude(source string, documentPath string) (*Document, []Diagnostic) {
	document, diagnostics := Parse(source)
	if document == nil {
		return document, diagnostics
	}

	root := document.Content()
	diagnostics = append(diagnostics, includePlacementDiagnostics(root)...)
	if root == nil || root.Kind != yaml.MappingNode {
		return document, diagnostics
	}

	commonNode := mappingValue(root, "common")
	if commonNode == nil || commonNode.Tag != includeTag {
		return document, diagnostics
	}

	includedNode, includeDiagnostics := resolveCommonInclude(commonNode, documentPath)
	diagnostics = append(diagnostics, includeDiagnostics...)
	if includedNode != nil {
		*commonNode = *includedNode
	}
	return document, diagnostics
}

func resolveCommonInclude(includeNode *yaml.Node, documentPath string) (*yaml.Node, []Diagnostic) {
	includePath := strings.TrimSpace(includeNode.Value)
	if includePath == "" {
		return nil, []Diagnostic{includeDiagnostic(includeNode, "common include path is empty")}
	}
	if filepath.IsAbs(includePath) {
		return nil, []Diagnostic{includeDiagnostic(includeNode, "common include path must be relative")}
	}
	if strings.TrimSpace(documentPath) == "" {
		return nil, []Diagnostic{includeDiagnostic(
			includeNode,
			fmt.Sprintf("common include %q cannot be resolved for an unsaved file", includePath),
		)}
	}

	resolvedPath := filepath.Clean(filepath.Join(filepath.Dir(documentPath), includePath))
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, []Diagnostic{includeDiagnostic(
			includeNode,
			fmt.Sprintf("common include %q cannot be read: %v", includePath, err),
		)}
	}

	includedDocument, diagnostics := Parse(string(content))
	if includedDocument == nil {
		return nil, includeDiagnosticsAtNode(includeNode, includePath, diagnostics)
	}

	includedRoot := includedDocument.Content()
	if hasIncludeTag(includedRoot) {
		return nil, []Diagnostic{includeDiagnostic(
			includeNode,
			fmt.Sprintf("common include %q contains nested !include, which is not supported", includePath),
		)}
	}
	return includedCommonNode(includedRoot), includeDiagnosticsAtNode(includeNode, includePath, diagnostics)
}

func includedCommonNode(root *yaml.Node) *yaml.Node {
	if commonNode := mappingValue(root, "common"); commonNode != nil {
		return commonNode
	}
	return root
}

func includePlacementDiagnostics(root *yaml.Node) []Diagnostic {
	var diagnostics []Diagnostic
	walkIncludeNodes(root, false, func(node *yaml.Node, allowed bool) {
		if node.Tag != includeTag || allowed {
			return
		}
		diagnostics = append(diagnostics, includeDiagnostic(node, "!include is only supported as the value of common"))
	})
	return diagnostics
}

func walkIncludeNodes(root *yaml.Node, allowed bool, visit func(*yaml.Node, bool)) {
	if root == nil {
		return
	}
	visit(root, allowed)
	if root.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(root.Content); i += 2 {
			keyNode := root.Content[i]
			valueNode := root.Content[i+1]
			walkIncludeNodes(valueNode, keyNode.Value == "common", visit)
		}
		return
	}
	for _, child := range root.Content {
		walkIncludeNodes(child, false, visit)
	}
}

func hasIncludeTag(root *yaml.Node) bool {
	if root == nil {
		return false
	}
	if root.Tag == includeTag {
		return true
	}
	for _, child := range root.Content {
		if hasIncludeTag(child) {
			return true
		}
	}
	return false
}

func includeDiagnosticsAtNode(node *yaml.Node, includePath string, diagnostics []Diagnostic) []Diagnostic {
	if len(diagnostics) == 0 {
		return nil
	}
	mapped := make([]Diagnostic, 0, len(diagnostics))
	position := NodePosition(node)
	for _, diagnostic := range diagnostics {
		mapped = append(mapped, Diagnostic{
			Message: fmt.Sprintf("common include %q: %s", includePath, diagnostic.Message),
			Line:    position.Line,
			Column:  position.Column,
		})
	}
	return mapped
}

func includeDiagnostic(node *yaml.Node, message string) Diagnostic {
	position := NodePosition(node)
	if position.Line <= 0 {
		position.Line = 1
	}
	if position.Column <= 0 {
		position.Column = 1
	}
	return Diagnostic{
		Message: message,
		Line:    position.Line,
		Column:  position.Column,
	}
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}
