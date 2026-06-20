package agent

import (
	"encoding/json"
	"strings"

	"cursor-byok/internal/relay"
)

// EstimatePromptTokens gives a rough prompt token budget for messages +
// tool definitions. Rough heuristic: ~4 bytes per token plus small
// per-message/per-tool overhead. Good enough for pre-flight truncation.
func EstimatePromptTokens(messages []openAIMessage, tools []openAITool) int {
	total := 0
	for _, m := range messages {
		total += 4
		total += messageTokenWeight(m)
	}
	for _, t := range tools {
		total += 40
		total += toolTokenWeight(t)
	}
	return total
}

func messageTokenWeight(m openAIMessage) int {
	contentBytes := rawMessageContentBytes(m)
	tokens := contentBytes / 4
	if tokens < 1 {
		tokens = 1
	}
	for _, tc := range m.ToolCalls {
		tokens += 16
		tokens += len(tc.Function.Name) / 4
		tokens += len(tc.Function.Arguments) / 4
	}
	return tokens
}

func rawMessageContentBytes(m openAIMessage) int {
	if len(m.Content) == 0 {
		return 0
	}
	s := string(m.Content)
	if m.Role == "tool" {
		s = strings.TrimSpace(s)
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
	}
	return len(s)
}

func toolTokenWeight(t openAITool) int {
	var fn struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Parameters  string `json:"parameters"`
	}
	if err := json.Unmarshal(t.Function, &fn); err != nil {
		return 0
	}
	return (len(fn.Name) + len(fn.Description) + len(fn.Parameters)) / 4
}

// AdapterContextWindow returns the effective prompt token budget for an
// adapter. Zero means unknown; caller should apply a conservative fallback.
func AdapterContextWindow(a relay.AdapterInfo) int {
	if a.ContextWindow > 0 {
		return a.ContextWindow
	}
	return 0
}

// TruncateMessages drops older history until estimated token usage fits the
// budget while keeping the current user message and recent context.
func TruncateMessages(messages []openAIMessage, tools []openAITool, budget int) []openAIMessage {
	if len(messages) == 0 || budget <= 0 {
		return messages
	}
	if EstimatePromptTokens(messages, tools) <= budget {
		return messages
	}

	// keep always includes the last user turn; remove oldest messages first.
	keep := messages
	for len(keep) > 1 {
		trimmed := keep[1:]
		if EstimatePromptTokens(trimmed, tools) <= budget {
			trimmed = append(trimmed, textMessage("user", "[Earlier conversation history truncated to fit context window. Continue from latest context.]"))
			return trimmed
		}
		keep = trimmed
	}
	return keep
}
