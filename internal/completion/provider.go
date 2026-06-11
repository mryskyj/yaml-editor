package completion

import (
	"sort"
	"strings"

	"github.com/mryskyj/yaml-editor/internal/schema"
	"gopkg.in/yaml.v3"
)

// Provide returns schema-aware YAML completion candidates at a cursor position.
func Provide(source string, line int, column int, root *schema.Field) []Candidate {
	return ProvideWithTools(source, line, column, root, nil)
}

// ProvideWithTools returns YAML completion candidates including tool-specific args schemas.
func ProvideWithTools(source string, line int, column int, root *schema.Field, toolSchemas map[string]*schema.Field) []Candidate {
	if root == nil || line <= 0 {
		return nil
	}

	lines := strings.Split(source, "\n")
	cursorIndex := min(line-1, len(lines)-1)
	cursorIndent := indentation(lineAt(lines, cursorIndex))
	path := inferPath(lines, cursorIndex, cursorIndent)
	if candidate := argsRootCollectionCandidate(path, lines, toolSchemas); candidate != nil {
		return []Candidate{*candidate}
	}
	current := fieldAtPathWithTools(root, path, lines, toolSchemas)
	if current == nil {
		return nil
	}

	if valueField := valueFieldAtCursor(lineAt(lines, cursorIndex), column, current); valueField != nil {
		if valueField.Name == "tool" {
			return toolCandidates(lineAt(lines, cursorIndex), column, toolSchemas)
		}
		if valueField.Name == "day_ref" {
			return dayRefCandidates(source, valueField)
		}
		if valueField.Name == "schedule_ref" {
			return scheduleRefCandidates(source, valueField)
		}
		return enumCandidates(valueField)
	}

	existing := existingKeys(lines, cursorIndex, cursorIndent, path)
	candidates := make([]Candidate, 0, len(current.Children))
	for _, child := range current.Children {
		if child == nil || existing[child.Name] {
			continue
		}
		candidates = append(candidates, candidateFromField(child))
	}

	return candidates
}

func dayRefCandidates(source string, field *schema.Field) []Candidate {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source), &document); err != nil || len(document.Content) == 0 {
		return nil
	}

	dates := mappingValue(mappingValue(document.Content[0], "common"), "dates")
	if dates == nil || dates.Kind != yaml.MappingNode {
		return nil
	}

	candidates := make([]Candidate, 0, len(dates.Content)/2)
	for index := 0; index+1 < len(dates.Content); index += 2 {
		keyNode := dates.Content[index]
		valueNode := dates.Content[index+1]
		if keyNode.Kind != yaml.ScalarNode || keyNode.Value == "" {
			continue
		}

		candidates = append(candidates, Candidate{
			Name:        keyNode.Value,
			Type:        field.Type,
			Description: dateRefDescription(valueNode),
			Required:    field.Required,
			Default:     field.Default,
			Enum:        field.Enum,
		})
	}
	return candidates
}

func dateRefDescription(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return ""
	}

	parts := make([]string, 0, 2)
	if date := mappingValue(node, "date"); date != nil && date.Kind == yaml.ScalarNode {
		parts = append(parts, "date: "+date.Value)
	}
	if holiday := mappingValue(node, "holiday"); holiday != nil && holiday.Kind == yaml.ScalarNode {
		parts = append(parts, "holiday: "+holiday.Value)
	}
	return strings.Join(parts, ", ")
}

func scheduleRefCandidates(source string, field *schema.Field) []Candidate {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source), &document); err != nil || len(document.Content) == 0 {
		return nil
	}

	schedules := mappingValue(mappingValue(document.Content[0], "common"), "schedules")
	if schedules == nil || schedules.Kind != yaml.MappingNode {
		return nil
	}

	candidates := make([]Candidate, 0, len(schedules.Content)/2)
	for index := 0; index+1 < len(schedules.Content); index += 2 {
		keyNode := schedules.Content[index]
		valueNode := schedules.Content[index+1]
		if keyNode.Kind != yaml.ScalarNode || keyNode.Value == "" {
			continue
		}

		candidates = append(candidates, Candidate{
			Name:        keyNode.Value,
			Type:        field.Type,
			Description: scheduleRefDescription(valueNode),
			Required:    field.Required,
			Default:     field.Default,
			Enum:        field.Enum,
		})
	}
	return candidates
}

func scheduleRefDescription(node *yaml.Node) string {
	if node == nil || node.Kind != yaml.ScalarNode {
		return ""
	}

	parts := make([]string, 0, 2)
	if node.Value != "" {
		parts = append(parts, "value: "+node.Value)
	}
	if node.LineComment != "" {
		parts = append(parts, node.LineComment)
	}
	return strings.Join(parts, ", ")
}

func toolCandidates(line string, column int, toolSchemas map[string]*schema.Field) []Candidate {
	if len(toolSchemas) == 0 {
		return nil
	}

	token := toolTokenAtCursor(line, column)
	prefix := strings.TrimSuffix(token, toolTokenLastSegment(token))
	segments := make(map[string]bool)
	for name := range toolSchemas {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		remainder := strings.TrimPrefix(name, prefix)
		if remainder == "" {
			continue
		}
		segment, _, hasRest := strings.Cut(remainder, ".")
		if segment == "" {
			continue
		}
		if hasRest {
			segments[segment+"."] = true
			continue
		}
		segments[segment] = true
	}

	names := make([]string, 0, len(segments))
	for name := range segments {
		names = append(names, name)
	}
	sort.Strings(names)

	return stringCandidates(names)
}

func toolTokenLastSegment(token string) string {
	if token == "" {
		return ""
	}
	index := strings.LastIndex(token, ".")
	if index < 0 {
		return token
	}
	return token[index+1:]
}

func stringCandidates(names []string) []Candidate {
	candidates := make([]Candidate, 0, len(names))
	for _, name := range names {
		candidates = append(candidates, Candidate{
			Name: name,
			Type: schema.FieldTypeString,
		})
	}
	return candidates
}

func toolTokenAtCursor(line string, column int) string {
	if column <= 0 {
		return ""
	}

	end := min(column-1, len(line))
	prefix := line[:end]
	colonIndex := strings.LastIndex(prefix, ":")
	if colonIndex < 0 {
		return ""
	}

	value := prefix[colonIndex+1:]
	value = strings.TrimLeft(value, " \t")
	value = strings.TrimLeft(value, `"'`)
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return value
	}
	return strings.Trim(fields[len(fields)-1], `"'`)
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

func candidateFromField(field *schema.Field) Candidate {
	candidate := Candidate{
		Name:        field.Name,
		Type:        field.Type,
		Description: field.Description,
		Required:    field.Required,
		Default:     field.Default,
		Enum:        field.Enum,
	}

	if len(field.Children) > 0 {
		candidate.Children = make([]Candidate, 0, len(field.Children))
		for _, child := range field.Children {
			if child == nil {
				continue
			}
			candidate.Children = append(candidate.Children, candidateFromField(child))
		}
	}
	if field.Item != nil {
		item := candidateFromField(field.Item)
		candidate.Item = &item
	}
	if field.MapValue != nil {
		mapValue := candidateFromField(field.MapValue)
		candidate.MapValue = &mapValue
	}

	return candidate
}

func argsRootCollectionCandidate(
	path []string,
	lines []string,
	toolSchemas map[string]*schema.Field,
) *Candidate {
	if len(toolSchemas) == 0 {
		return nil
	}
	argsIndex := lastPathIndex(path, "args")
	if argsIndex < 0 || len(path[argsIndex+1:]) != 0 {
		return nil
	}
	toolName := toolValueAtPath(lines, path[:argsIndex])
	toolSchema := toolSchemas[toolName]
	if toolSchema == nil {
		return nil
	}
	if toolSchema.Type != schema.FieldTypeSlice && toolSchema.Type != schema.FieldTypeArray {
		return nil
	}
	candidate := candidateFromField(toolSchema)
	candidate.Name = ""
	candidate.Root = true
	return &candidate
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
	return fieldAtPathFrom(root, path)
}

func fieldAtPathWithTools(
	root *schema.Field,
	path []string,
	lines []string,
	toolSchemas map[string]*schema.Field,
) *schema.Field {
	argsIndex := lastPathIndex(path, "args")
	if argsIndex >= 0 {
		toolName := toolValueAtPath(lines, path[:argsIndex])
		if toolSchema := toolSchemas[toolName]; toolSchema != nil {
			return fieldAtPathFrom(toolSchema, path[argsIndex+1:])
		}
	}
	return fieldAtPathFrom(root, path)
}

func fieldAtPathFrom(root *schema.Field, path []string) *schema.Field {
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

func lastPathIndex(path []string, name string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == name {
			return i
		}
	}
	return -1
}

func toolValueAtPath(lines []string, parentPath []string) string {
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, ok := yamlKey(line)
		if !ok || key != "tool" {
			continue
		}
		indent := keyIndentation(line)
		linePath := append(inferPath(lines, i, indent), key)
		if samePath(linePath, appendPath(parentPath, "tool")) {
			return yamlValue(line)
		}
	}
	return ""
}

func appendPath(path []string, value string) []string {
	next := make([]string, 0, len(path)+1)
	next = append(next, path...)
	next = append(next, value)
	return next
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

func yamlValue(line string) string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "- ")
	index := strings.Index(trimmed, ":")
	if index < 0 {
		return ""
	}

	value := strings.TrimSpace(trimmed[index+1:])
	value = strings.Trim(value, `"'`)
	return value
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
