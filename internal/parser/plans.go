package parser

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Plan represents a parsed implementation plan
type Plan struct {
	Name      string    `json:"name"`
	Title     string    `json:"title,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ParsePlan parses a markdown plan file
func ParsePlan(path string) (*Plan, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Get file info for created_at
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Extract name from filename (without .md extension)
	name := filepath.Base(strings.TrimSuffix(path, ".md"))

	// Extract title from first heading
	title := name
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			title = strings.TrimSpace(title)
			break
		}
	}

	return &Plan{
		Name:      name,
		Title:     title,
		Content:   string(content),
		CreatedAt: info.ModTime(),
	}, nil
}
