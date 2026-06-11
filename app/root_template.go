package app

import (
	"strconv"
	"strings"

	"github.com/mryskyj/yaml-editor/internal/schema"
	"gopkg.in/yaml.v3"
)

func rootSchemaTemplate(root *schema.Field, defaults *yaml.Node) string {
	if root == nil || root.Type != schema.FieldTypeStruct {
		return ""
	}

	var lines []string
	for _, child := range requiredChildren(root) {
		appendFieldTemplate(&lines, child, 0, "", mappingValue(defaults, child.Name))
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func rootScheduleTemplate(root *schema.Field, defaults *yaml.Node) string {
	common, ok := root.FindChild("common")
	if !ok {
		return ""
	}
	schedules, ok := common.FindChild("schedules")
	if !ok {
		return ""
	}
	defaultNode := mappingValue(mappingValue(defaults, "common"), "schedules")
	var lines []string
	appendMapEntries(&lines, schedules, 0, defaultNode)
	return strings.Join(lines, "\n")
}

func loadBuiltinRootDefaults() *yaml.Node {
	var document yaml.Node
	if err := yaml.Unmarshal(rootDefaultsSource, &document); err != nil {
		return nil
	}
	if len(document.Content) == 0 {
		return nil
	}
	return document.Content[0]
}

func appendFieldTemplate(lines *[]string, field *schema.Field, indent int, prefix string, defaultNode *yaml.Node) {
	if field == nil || field.Name == "" {
		return
	}

	linePrefix := strings.Repeat(" ", indent) + prefix
	switch field.Type {
	case schema.FieldTypeStruct:
		*lines = append(*lines, linePrefix+field.Name+":")
		for _, child := range requiredChildren(field) {
			appendFieldTemplate(lines, child, indent+2, "", mappingValue(defaultNode, child.Name))
		}
	case schema.FieldTypeSlice, schema.FieldTypeArray:
		*lines = append(*lines, linePrefix+field.Name+":")
		appendListItemTemplate(lines, field.Item, indent+2, firstSequenceItem(defaultNode))
	case schema.FieldTypeMap:
		*lines = append(*lines, linePrefix+field.Name+":")
		appendMapEntries(lines, field, indent+2, defaultNode)
	default:
		*lines = append(*lines, linePrefix+field.Name+": "+scalarTemplateValue(field, defaultNode))
	}
}

func appendMapEntries(lines *[]string, field *schema.Field, indent int, defaultNode *yaml.Node) {
	if field == nil || field.MapValue == nil || defaultNode == nil || defaultNode.Kind != yaml.MappingNode {
		if field != nil && field.Name != "" && len(*lines) > 0 {
			(*lines)[len(*lines)-1] += " {}"
		}
		return
	}

	for index := 0; index+1 < len(defaultNode.Content); index += 2 {
		keyNode := defaultNode.Content[index]
		valueNode := defaultNode.Content[index+1]
		keyPrefix := strings.Repeat(" ", indent) + keyNode.Value

		switch field.MapValue.Type {
		case schema.FieldTypeStruct:
			*lines = append(*lines, keyPrefix+":")
			for _, child := range requiredChildren(field.MapValue) {
				appendFieldTemplate(lines, child, indent+2, "", mappingValue(valueNode, child.Name))
			}
		case schema.FieldTypeSlice, schema.FieldTypeArray:
			*lines = append(*lines, keyPrefix+":")
			appendListItemTemplate(lines, field.MapValue.Item, indent+2, firstSequenceItem(valueNode))
		case schema.FieldTypeMap:
			*lines = append(*lines, keyPrefix+":")
			appendMapEntries(lines, field.MapValue, indent+2, valueNode)
		default:
			*lines = append(*lines, keyPrefix+": "+scalarTemplateValue(field.MapValue, valueNode))
		}
	}
}

func appendListItemTemplate(lines *[]string, item *schema.Field, indent int, defaultNode *yaml.Node) {
	if item == nil {
		*lines = append(*lines, strings.Repeat(" ", indent)+"-")
		return
	}

	if item.Type != schema.FieldTypeStruct {
		*lines = append(*lines, strings.Repeat(" ", indent)+"- "+scalarTemplateValue(item, defaultNode))
		return
	}

	children := requiredChildren(item)
	if len(children) == 0 {
		*lines = append(*lines, strings.Repeat(" ", indent)+"- {}")
		return
	}

	appendFieldTemplate(lines, children[0], indent, "- ", mappingValue(defaultNode, children[0].Name))
	for _, child := range children[1:] {
		appendFieldTemplate(lines, child, indent+2, "", mappingValue(defaultNode, child.Name))
	}
}

func requiredChildren(field *schema.Field) []*schema.Field {
	if field == nil {
		return nil
	}

	children := make([]*schema.Field, 0, len(field.Children))
	for _, child := range field.Children {
		if child != nil && child.Required {
			children = append(children, child)
		}
	}
	return children
}

func scalarTemplateValue(field *schema.Field, defaultNode *yaml.Node) string {
	if field == nil {
		return "null"
	}
	if defaultNode != nil && defaultNode.Kind == yaml.ScalarNode {
		return scalarNodeLiteral(field.Type, defaultNode)
	}
	if field.Default != "" {
		return scalarLiteral(field.Type, field.Default)
	}
	if len(field.Enum) > 0 {
		return scalarLiteral(field.Type, field.Enum[0])
	}

	switch field.Type {
	case schema.FieldTypeString:
		return strconv.Quote("")
	case schema.FieldTypeBool:
		return "false"
	case schema.FieldTypeInt, schema.FieldTypeFloat:
		return "0"
	default:
		return "null"
	}
}

func scalarLiteral(fieldType schema.FieldType, value string) string {
	switch fieldType {
	case schema.FieldTypeString:
		return strconv.Quote(value)
	case schema.FieldTypeBool, schema.FieldTypeInt, schema.FieldTypeFloat:
		return value
	default:
		return strconv.Quote(value)
	}
}

func scalarNodeLiteral(fieldType schema.FieldType, node *yaml.Node) string {
	prefix := ""
	if node.Anchor != "" {
		prefix = "&" + node.Anchor + " "
	}
	suffix := ""
	if node.LineComment != "" {
		suffix = " " + node.LineComment
	}
	return prefix + scalarLiteral(fieldType, node.Value) + suffix
}

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		if node.Content[index].Value == key {
			return node.Content[index+1]
		}
	}
	return nil
}

func firstSequenceItem(node *yaml.Node) *yaml.Node {
	if node == nil || node.Kind != yaml.SequenceNode || len(node.Content) == 0 {
		return nil
	}
	return node.Content[0]
}
