package file

import (
	"fmt"
	"os"
	"unicode/utf8"
)

// Document represents an opened or newly created YAML document.
type Document struct {
	Path    string
	Content string
}

// Service provides UTF-8 text file operations.
type Service struct {
	recent *RecentStore
}

// NewService creates a file service.
func NewService(recent *RecentStore) *Service {
	return &Service{recent: recent}
}

// NewDocument returns an empty unsaved document.
func (s *Service) NewDocument() Document {
	return Document{}
}

// Open reads a UTF-8 text file and records it as recently opened.
func (s *Service) Open(path string) (Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("open YAML file: %w", err)
	}
	if !utf8.Valid(content) {
		return Document{}, fmt.Errorf("open YAML file: invalid UTF-8")
	}

	if s != nil && s.recent != nil {
		if err := s.recent.Add(path); err != nil {
			return Document{}, err
		}
	}

	return Document{
		Path:    path,
		Content: string(content),
	}, nil
}

// Save writes UTF-8 text content and records the path as recently opened.
func (s *Service) Save(path string, content string) error {
	if !utf8.ValidString(content) {
		return fmt.Errorf("save YAML file: invalid UTF-8")
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("save YAML file: %w", err)
	}

	if s != nil && s.recent != nil {
		if err := s.recent.Add(path); err != nil {
			return err
		}
	}

	return nil
}

// RecentFiles returns recently opened files.
func (s *Service) RecentFiles() ([]string, error) {
	if s == nil || s.recent == nil {
		return nil, nil
	}

	return s.recent.List()
}
