package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const defaultRecentLimit = 10

// RecentStore persists recently opened files as JSON.
type RecentStore struct {
	path  string
	limit int
}

// NewRecentStore creates a recent-file store backed by a JSON file.
func NewRecentStore(path string, limit int) *RecentStore {
	if limit <= 0 {
		limit = defaultRecentLimit
	}
	return &RecentStore{
		path:  path,
		limit: limit,
	}
}

// Add records a path as most recently opened.
func (s *RecentStore) Add(path string) error {
	if s == nil {
		return fmt.Errorf("recent store is nil")
	}
	if path == "" {
		return fmt.Errorf("recent file path is empty")
	}

	files, err := s.List()
	if err != nil {
		return err
	}

	next := make([]string, 0, len(files)+1)
	next = append(next, path)
	for _, file := range files {
		if file != path {
			next = append(next, file)
		}
	}
	if len(next) > s.limit {
		next = next[:s.limit]
	}

	return s.save(next)
}

// List returns recent files ordered from newest to oldest.
func (s *RecentStore) List() ([]string, error) {
	if s == nil {
		return nil, fmt.Errorf("recent store is nil")
	}
	if s.path == "" {
		return nil, fmt.Errorf("recent store path is empty")
	}

	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read recent files: %w", err)
	}

	var files []string
	if err := json.Unmarshal(content, &files); err != nil {
		return nil, fmt.Errorf("read recent files: %w", err)
	}
	return files, nil
}

func (s *RecentStore) save(files []string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("save recent files: %w", err)
	}

	content, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return fmt.Errorf("save recent files: %w", err)
	}

	if err := os.WriteFile(s.path, content, 0o644); err != nil {
		return fmt.Errorf("save recent files: %w", err)
	}
	return nil
}
