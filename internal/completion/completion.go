package completion

import "github.com/mryskyj/yaml-editor/internal/schema"

// Candidate describes a YAML key completion candidate.
type Candidate struct {
	Name        string
	Type        schema.FieldType
	Description string
	Required    bool
	Default     string
	Enum        []string
	Root        bool
	Children    []Candidate
	Item        *Candidate
	MapValue    *Candidate
}
