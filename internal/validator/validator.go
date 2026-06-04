package validator

import (
	"fmt"

	"github.com/mryskyj/yaml-editor/internal/schema"
	"github.com/mryskyj/yaml-editor/internal/yamlx"
	"gopkg.in/yaml.v3"
)

// Validate parses YAML source and compares it with the provided root schema.
func Validate(source string, root *schema.Field) []Diagnostic {
	document, yamlDiagnostics := yamlx.Parse(source)
	diagnostics := fromYAMLDiagnostics(yamlDiagnostics)
	if document == nil {
		return diagnostics
	}
	if root == nil {
		return append(diagnostics, newDiagnostic("root schema is not registered", 1, 1))
	}

	return append(diagnostics, validateNode(document.Content(), root)...)
}

func validateNode(node *yaml.Node, field *schema.Field) []Diagnostic {
	if field == nil || node == nil {
		return nil
	}

	if !matchesKind(node, field) {
		return []Diagnostic{nodeDiagnostic(
			node,
			fmt.Sprintf("key %q must be %s", field.Name, field.Type),
		)}
	}

	switch field.Type {
	case schema.FieldTypeStruct:
		return validateStruct(node, field)
	case schema.FieldTypeSlice, schema.FieldTypeArray:
		return validateSequence(node, field.Item)
	case schema.FieldTypeMap:
		return validateMap(node, field.MapValue)
	default:
		return nil
	}
}

func validateStruct(node *yaml.Node, field *schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		child, ok := field.FindChild(keyNode.Value)
		if !ok {
			diagnostics = append(diagnostics, nodeDiagnostic(
				keyNode,
				fmt.Sprintf("undefined key %q", keyNode.Value),
			))
			continue
		}

		diagnostics = append(diagnostics, validateNode(valueNode, child)...)
	}

	return diagnostics
}

func validateSequence(node *yaml.Node, item *schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	for _, child := range node.Content {
		diagnostics = append(diagnostics, validateNode(child, item)...)
	}

	return diagnostics
}

func validateMap(node *yaml.Node, value *schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	for i := 1; i < len(node.Content); i += 2 {
		diagnostics = append(diagnostics, validateNode(node.Content[i], value)...)
	}

	return diagnostics
}

func matchesKind(node *yaml.Node, field *schema.Field) bool {
	switch field.Type {
	case schema.FieldTypeStruct, schema.FieldTypeMap:
		return node.Kind == yaml.MappingNode
	case schema.FieldTypeSlice, schema.FieldTypeArray:
		return node.Kind == yaml.SequenceNode
	case schema.FieldTypeString:
		return node.Kind == yaml.ScalarNode && node.Tag == "!!str"
	case schema.FieldTypeBool:
		return node.Kind == yaml.ScalarNode && node.Tag == "!!bool"
	case schema.FieldTypeInt:
		return node.Kind == yaml.ScalarNode && node.Tag == "!!int"
	case schema.FieldTypeFloat:
		return node.Kind == yaml.ScalarNode && (node.Tag == "!!float" || node.Tag == "!!int")
	default:
		return false
	}
}

func fromYAMLDiagnostics(yamlDiagnostics []yamlx.Diagnostic) []Diagnostic {
	diagnostics := make([]Diagnostic, 0, len(yamlDiagnostics))
	for _, diagnostic := range yamlDiagnostics {
		diagnostics = append(diagnostics, newDiagnostic(
			diagnostic.Message,
			diagnostic.Line,
			diagnostic.Column,
		))
	}

	return diagnostics
}

func nodeDiagnostic(node *yaml.Node, message string) Diagnostic {
	position := yamlx.NodePosition(node)
	return newDiagnostic(message, position.Line, position.Column)
}

func newDiagnostic(message string, line int, column int) Diagnostic {
	if line <= 0 {
		line = 1
	}
	if column <= 0 {
		column = 1
	}

	return Diagnostic{
		Severity:  SeverityError,
		Message:   message,
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + 1,
	}
}
