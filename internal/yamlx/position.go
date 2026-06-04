package yamlx

import "gopkg.in/yaml.v3"

// Position represents a 1-based editor position.
type Position struct {
	Line   int
	Column int
}

// NodePosition returns the source position recorded on a YAML node.
func NodePosition(node *yaml.Node) Position {
	if node == nil {
		return Position{}
	}

	return Position{
		Line:   node.Line,
		Column: node.Column,
	}
}
