package filter

import (
	"path/filepath"
	"strings"

	"github.com/dkd/claude-insights-agent/internal/config"
	"github.com/dkd/claude-insights-agent/internal/parser"
)

// Filter applies privacy rules to sessions based on config
type Filter struct {
	cfg *config.SharingConfig
}

// New creates a new Filter with the given config
func New(cfg *config.SharingConfig) *Filter {
	return &Filter{cfg: cfg}
}

// Apply filters a session according to share level settings
func (f *Filter) Apply(s *parser.Session) *parser.Session {
	// Check if project is excluded
	if f.isExcluded(s.ProjectPath) {
		return nil
	}

	// Create filtered copy
	filtered := &parser.Session{
		ID:             s.ID,
		StartedAt:      s.StartedAt,
		EndedAt:        s.EndedAt,
		TotalMessages:  s.TotalMessages,
		TotalTokensIn:  s.TotalTokensIn,
		TotalTokensOut: s.TotalTokensOut,
		Model:          s.Model,
		Tools:          s.Tools,
		Tags:           s.Tags,
	}

	// Anonymize or include project name
	if f.cfg.AnonymizePaths {
		filtered.ProjectName = filepath.Base(s.ProjectPath)
	} else {
		filtered.ProjectName = s.ProjectPath
	}

	// Apply share level
	switch f.cfg.Level {
	case "none":
		// Don't share anything
		return nil

	case "metadata":
		// Share metadata only, no message content
		filtered.Messages = nil

	case "full":
		// Share everything including messages
		filtered.Messages = s.Messages
	}

	return filtered
}

// isExcluded checks if project matches any exclusion pattern
func (f *Filter) isExcluded(projectPath string) bool {
	for _, pattern := range f.cfg.ExcludeProjects {
		matched, _ := filepath.Match(pattern, projectPath)
		if matched {
			return true
		}
		// Also check if pattern matches any part of the path
		if strings.Contains(pattern, "**") {
			// Simple glob matching for **
			cleanPattern := strings.ReplaceAll(pattern, "**", "*")
			parts := strings.Split(projectPath, "/")
			for _, part := range parts {
				if matched, _ := filepath.Match(cleanPattern, part); matched {
					return true
				}
			}
		}
	}
	return false
}
