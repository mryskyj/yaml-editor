package completion

import (
	"strings"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

// Provide returns schema-aware YAML completion candidates at a cursor position.
func Provide(source string, line int, column int, root *schema.Field) []Candidate {
	if root == nil || line <= 0 {
		return nil
	}

	lines := strings.Split(source, "\n")
	cursorIndex := min(line-1, len(lines)-1)
	cursorIndent := indentation(lineAt(lines, cursorIndex))
	path := inferPath(lines, cursorIndex, cursorIndent)
	current := fieldAtPath(root, path)
	if current == nil {
		return nil
	}

	if valueField := valueFieldAtCursor(lineAt(lines, cursorIndex), column, current); valueField != nil {
		return enumCandidates(valueField)
	}

	existing := existingKeys(lines, cursorIndex, cursorIndent, path)
	candidates := make([]Candidate, 0, len(current.Children))
	for _, child := range current.Children {
		if child == nil || existing[child.Name] {
			continue
		}
		candidates = append(candidates, Candidate{
			Name:        child.Name,
			Type:        child.Type,
			Description: child.Description,
			Required:    child.Required,
			Default:     child.Default,
			Enum:        child.Enum,
		})
	}

	return candidates
}

func valueFieldAtCursor(line string, column int, current *schema.Field) *schema.Field {
	if current == nil || column <= 0 {
		return nil
	}

	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	colonIndex := strings.Index(trimmed, ":")
	if colonIndex <= 0 {
		return nil
	}

	lineColonColumn := indentation(line) + colonIndex + 1
	if column <= lineColonColumn {
		return nil
	}

	child, ok := current.FindChild(strings.TrimSpace(trimmed[:colonIndex]))
	if !ok {
		return nil
	}
	return child
}

func enumCandidates(field *schema.Field) []Candidate {
	if field == nil || len(field.Enum) == 0 {
		return nil
	}

	candidates := make([]Candidate, 0, len(field.Enum))
	for _, value := range field.Enum {
		candidates = append(candidates, Candidate{
			Name:        value,
			Type:        field.Type,
			Description: field.Description,
			Required:    field.Required,
			Default:     field.Default,
			Enum:        field.Enum,
		})
	}
	return candidates
}

func inferPath(lines []string, cursorIndex int, cursorIndent int) []string {
	stack := make([]pathEntry, 0)
	for i := 0; i < cursorIndex && i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := keyIndentation(line)
		if indent >= cursorIndent {
			continue
		}

		key, hasKey := yamlContainerKey(line)
		if !hasKey {
			continue
		}

		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		stack = append(stack, pathEntry{indent: indent, key: key})
	}

	for len(stack) > 0 && stack[len(stack)-1].indent >= cursorIndent {
		stack = stack[:len(stack)-1]
	}

	path := make([]string, 0, len(stack))
	for _, entry := range stack {
		path = append(path, entry.key)
	}
	return path
}

type pathEntry struct {
	indent int
	key    string
}

func fieldAtPath(root *schema.Field, path []string) *schema.Field {
	current := root
	for _, name := range path {
		current = collectionValueField(current)
		if current == nil {
			return nil
		}

		child, ok := current.FindChild(name)
		if !ok {
			return nil
		}
		current = child
	}
	return collectionValueField(current)
}

func collectionValueField(field *schema.Field) *schema.Field {
	if field == nil {
		return nil
	}

	switch field.Type {
	case schema.FieldTypeSlice, schema.FieldTypeArray:
		if field.Item != nil {
			return field.Item
		}
	case schema.FieldTypeMap:
		if field.MapValue != nil {
			return field.MapValue
		}
	}
	return field
}

func existingKeys(lines []string, cursorIndex int, cursorIndent int, path []string) map[string]bool {
	keys := make(map[string]bool)
	for i := 0; i < len(lines); i++ {
		if i == cursorIndex {
			continue
		}
		line := lines[i]
		if strings.TrimSpace(line) == "" || keyIndentation(line) != cursorIndent {
			continue
		}

		linePath := inferPath(lines, i, cursorIndent)
		if !samePath(linePath, path) {
			continue
		}

		key, hasKey := yamlKey(line)
		if hasKey {
			keys[key] = true
		}
	}

	return keys
}

func yamlKey(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	index := strings.Index(trimmed, ":")
	if index <= 0 {
		return "", false
	}

	return strings.TrimSpace(trimmed[:index]), true
}

func yamlContainerKey(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	index := strings.Index(trimmed, ":")
	if index <= 0 {
		return "", false
	}

	if strings.TrimSpace(trimmed[index+1:]) != "" {
		return "", false
	}
	return strings.TrimSpace(trimmed[:index]), true
}

func indentation(line string) int {
	count := 0
	for _, char := range line {
		if char != ' ' {
			break
		}
		count++
	}
	return count
}

func keyIndentation(line string) int {
	indent := indentation(line)
	if strings.HasPrefix(strings.TrimSpace(line), "- ") {
		return indent + 2
	}
	return indent
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func samePath(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
