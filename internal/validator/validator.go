package validator

import (
	"fmt"
	"strings"

	"github.com/mryskyj/yaml-editor/internal/schema"
	"github.com/mryskyj/yaml-editor/internal/yamlx"
	"gopkg.in/yaml.v3"
)

// Validate parses YAML source and compares it with the provided root schema.
func Validate(source string, root *schema.Field) []Diagnostic {
	return ValidateWithTools(source, root, nil)
}

// ValidateWithTools parses YAML source and compares it with root plus tool-specific args schemas.
func ValidateWithTools(source string, root *schema.Field, toolSchemas map[string]*schema.Field) []Diagnostic {
	document, yamlDiagnostics := yamlx.Parse(source)
	diagnostics := fromYAMLDiagnostics(yamlDiagnostics)
	if document == nil {
		return diagnostics
	}
	if root == nil {
		return append(diagnostics, newDiagnostic("root schema is not registered", 1, 1))
	}

	return append(diagnostics, validateNode(document.Content(), root, toolSchemas)...)
}

func validateNode(node *yaml.Node, field *schema.Field, toolSchemas map[string]*schema.Field) []Diagnostic {
	if field == nil || node == nil {
		return nil
	}

	if !matchesKind(node, field) {
		return []Diagnostic{nodeDiagnostic(
			node,
			fmt.Sprintf("key %q must be %s", field.Name, field.Type),
		)}
	}

	if len(field.Enum) > 0 && !contains(field.Enum, node.Value) {
		return []Diagnostic{nodeDiagnostic(
			node,
			fmt.Sprintf("key %q must be one of: %s", field.Name, strings.Join(field.Enum, ", ")),
		)}
	}

	switch field.Type {
	case schema.FieldTypeStruct:
		return validateStruct(node, field, toolSchemas)
	case schema.FieldTypeSlice, schema.FieldTypeArray:
		return validateSequence(node, field.Item, toolSchemas)
	case schema.FieldTypeMap:
		return validateMap(node, field.MapValue, toolSchemas)
	default:
		return nil
	}
}

func validateStruct(node *yaml.Node, field *schema.Field, toolSchemas map[string]*schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	seen := make(map[string]bool, len(node.Content)/2)
	toolName, toolNode := mappingStringValue(node, "tool")
	toolSchema := toolSchemas[toolName]
	for i := 0; i+1 < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		seen[keyNode.Value] = true

		if keyNode.Value == "tool" && toolName != "" && len(toolSchemas) > 0 && toolSchema == nil {
			diagnostics = append(diagnostics, nodeDiagnostic(
				valueNode,
				fmt.Sprintf("tool %q is not registered", toolName),
			))
		}

		child, ok := field.FindChild(keyNode.Value)
		if !ok {
			diagnostics = append(diagnostics, nodeDiagnostic(
				keyNode,
				fmt.Sprintf("undefined key %q", keyNode.Value),
			))
			continue
		}

		if keyNode.Value == "args" && toolSchema != nil {
			diagnostics = append(diagnostics, validateNode(valueNode, toolSchema, toolSchemas)...)
			continue
		}

		diagnostics = append(diagnostics, validateNode(valueNode, child, toolSchemas)...)
	}

	if toolNode != nil && toolName == "" && len(toolSchemas) > 0 {
		diagnostics = append(diagnostics, nodeDiagnostic(toolNode, "tool must not be empty"))
	}

	for _, child := range field.Children {
		if child != nil && child.Required && !seen[child.Name] {
			diagnostics = append(diagnostics, nodeDiagnostic(
				node,
				fmt.Sprintf("required key %q is missing", child.Name),
			))
		}
	}

	return diagnostics
}

func validateSequence(node *yaml.Node, item *schema.Field, toolSchemas map[string]*schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	for _, child := range node.Content {
		diagnostics = append(diagnostics, validateNode(child, item, toolSchemas)...)
	}

	return diagnostics
}

func validateMap(node *yaml.Node, value *schema.Field, toolSchemas map[string]*schema.Field) []Diagnostic {
	var diagnostics []Diagnostic
	for i := 1; i < len(node.Content); i += 2 {
		diagnostics = append(diagnostics, validateNode(node.Content[i], value, toolSchemas)...)
	}

	return diagnostics
}

func mappingStringValue(node *yaml.Node, key string) (string, *yaml.Node) {
	if node == nil || node.Kind != yaml.MappingNode {
		return "", nil
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != key {
			continue
		}
		valueNode := node.Content[i+1]
		return strings.TrimSpace(valueNode.Value), valueNode
	}
	return "", nil
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

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
