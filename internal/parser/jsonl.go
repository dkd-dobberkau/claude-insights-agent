package parser

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Session represents a parsed Claude Code session
type Session struct {
	ID             string               `json:"session_id"`
	ProjectName    string               `json:"project_name"`
	ProjectPath    string               `json:"-"` // Not sent to server
	StartedAt      time.Time            `json:"started_at"`
	EndedAt        *time.Time           `json:"ended_at,omitempty"`
	TotalMessages  int                  `json:"total_messages"`
	TotalTokensIn  int                  `json:"total_tokens_in"`
	TotalTokensOut int                  `json:"total_tokens_out"`
	Model          string               `json:"model,omitempty"`
	Tools          map[string]*ToolStats `json:"tools"`
	Tags           []string             `json:"tags"`
	Messages       []Message            `json:"messages,omitempty"`
	TokenUsage     []TokenUsageItem     `json:"token_usage"`
	ToolCalls      []ToolCallItem       `json:"tool_calls"`
}

type ToolStats struct {
	Count   int `json:"count"`
	Success int `json:"success"`
	Errors  int `json:"errors"`
}

type Message struct {
	Seq       int       `json:"seq"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
}

// TokenUsageItem represents per-message token usage
type TokenUsageItem struct {
	MessageSeq          int       `json:"message_sequence,omitempty"`
	Timestamp           time.Time `json:"timestamp,omitempty"`
	Model               string    `json:"model,omitempty"`
	InputTokens         int       `json:"input_tokens"`
	OutputTokens        int       `json:"output_tokens"`
	CacheReadTokens     int       `json:"cache_read_tokens"`
	CacheCreationTokens int       `json:"cache_creation_tokens"`
}

// ToolCallItem represents a detailed tool call
type ToolCallItem struct {
	MessageSeq int    `json:"message_sequence,omitempty"`
	ToolName   string `json:"tool_name"`
	ToolInput  string `json:"tool_input,omitempty"`
	ToolOutput string `json:"tool_output,omitempty"`
	DurationMs int    `json:"duration_ms,omitempty"`
	Success    bool   `json:"success"`
}

// RawEntry represents a single line in JSONL
type RawEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp,omitempty"`
	Message   json.RawMessage `json:"message,omitempty"`
	Role      string          `json:"role,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
}

type MessageContent struct {
	Content []ContentBlock `json:"content"`
	Usage   *Usage         `json:"usage,omitempty"`
	Model   string         `json:"model,omitempty"`
}

type ContentBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// ParseJSONL parses a JSONL session file
func ParseJSONL(path string) (*Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	session := &Session{
		ID:         filepath.Base(strings.TrimSuffix(path, ".jsonl")),
		Tools:      make(map[string]*ToolStats),
		Tags:       []string{},
		TokenUsage: []TokenUsageItem{},
		ToolCalls:  []ToolCallItem{},
	}

	// Extract project path from parent directory
	parentDir := filepath.Base(filepath.Dir(path))
	if strings.HasPrefix(parentDir, "-") {
		session.ProjectPath = strings.ReplaceAll(parentDir, "-", "/")
		parts := strings.Split(session.ProjectPath, "/")
		if len(parts) > 0 {
			session.ProjectName = parts[len(parts)-1]
		}
	}

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var firstTs, lastTs time.Time
	msgSeq := 0

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry RawEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // Skip invalid lines
		}

		// Parse timestamp
		if entry.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				if firstTs.IsZero() {
					firstTs = ts
				}
				lastTs = ts
			}
		}

		// Handle different entry types
		switch entry.Type {
		case "user", "assistant":
			session.TotalMessages++

			var msgContent MessageContent
			if entry.Message != nil {
				json.Unmarshal(entry.Message, &msgContent)
			}

			// Parse message timestamp
			msgTs, _ := time.Parse(time.RFC3339, entry.Timestamp)

			// Extract text and tool usage
			var textParts []string
			for _, block := range msgContent.Content {
				switch block.Type {
				case "text":
					textParts = append(textParts, block.Text)
				case "tool_use":
					if block.Name != "" {
						if session.Tools[block.Name] == nil {
							session.Tools[block.Name] = &ToolStats{}
						}
						session.Tools[block.Name].Count++
						session.Tools[block.Name].Success++ // Assume success

						// Collect detailed tool call
						var toolInput string
						if block.Input != nil {
							if inputBytes, err := json.Marshal(block.Input); err == nil {
								toolInput = string(inputBytes)
							}
						}
						session.ToolCalls = append(session.ToolCalls, ToolCallItem{
							MessageSeq: msgSeq,
							ToolName:   block.Name,
							ToolInput:  toolInput,
							Success:    true,
						})
					}
				}
			}

			// Track tokens and model
			if msgContent.Usage != nil {
				session.TotalTokensIn += msgContent.Usage.InputTokens
				session.TotalTokensOut += msgContent.Usage.OutputTokens

				// Collect detailed token usage per message
				session.TokenUsage = append(session.TokenUsage, TokenUsageItem{
					MessageSeq:          msgSeq,
					Timestamp:           msgTs,
					Model:               msgContent.Model,
					InputTokens:         msgContent.Usage.InputTokens,
					OutputTokens:        msgContent.Usage.OutputTokens,
					CacheReadTokens:     msgContent.Usage.CacheReadInputTokens,
					CacheCreationTokens: msgContent.Usage.CacheCreationInputTokens,
				})
			}
			if msgContent.Model != "" {
				session.Model = msgContent.Model
			}

			// Store message
			ts := msgTs
			session.Messages = append(session.Messages, Message{
				Seq:       msgSeq,
				Timestamp: ts,
				Role:      entry.Type,
				Content:   strings.Join(textParts, "\n"),
			})
			msgSeq++
		}
	}

	if !firstTs.IsZero() {
		session.StartedAt = firstTs
	}
	if !lastTs.IsZero() {
		session.EndedAt = &lastTs
	}

	// Auto-generate tags
	session.Tags = generateTags(session)

	return session, scanner.Err()
}

func generateTags(s *Session) []string {
	tags := make([]string, 0)

	// Tag by tools used
	for tool := range s.Tools {
		tags = append(tags, "tool:"+tool)
	}

	// Simple content-based tags (check first few messages)
	content := ""
	for i, msg := range s.Messages {
		if i >= 5 {
			break
		}
		content += strings.ToLower(msg.Content) + " "
	}

	patterns := map[string][]string{
		"debugging":     {"error", "bug", "fix", "debug"},
		"refactoring":   {"refactor", "cleanup", "restructure"},
		"feature":       {"implement", "add feature", "new feature"},
		"testing":       {"test", "spec", "coverage"},
		"documentation": {"document", "readme", "comment"},
	}

	for tag, keywords := range patterns {
		for _, kw := range keywords {
			if strings.Contains(content, kw) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}
