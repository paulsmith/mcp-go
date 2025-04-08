package mcp

import (
	"context"
	"encoding/json"
)

// PromptHandler is a function that handles prompt requests
type PromptHandler func(ctx context.Context, args map[string]interface{}) ([]PromptMessage, error)

// AddPrompt registers a prompt with the server
func (s *Server) AddPrompt(name, description string, arguments []PromptArgument, handler PromptHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prompt := Prompt{
		Name:        name,
		Description: description,
		Arguments:   arguments,
	}

	// Register the prompt
	s.prompts = append(s.prompts, prompt)
	s.promptHandlers[name] = handler
}

// handleListPrompts handles a prompts/list request
func (s *Server) handleListPrompts(ctx context.Context, msg *Message) {
	s.mu.RLock()
	prompts := make([]Prompt, len(s.prompts))
	copy(prompts, s.prompts)
	s.mu.RUnlock()

	// Build response
	result := struct {
		Prompts []Prompt `json:"prompts"`
	}{
		Prompts: prompts,
	}

	s.sendResult(ctx, msg.ID, result)
}

// handleGetPrompt handles a prompts/get request
func (s *Server) handleGetPrompt(ctx context.Context, msg *Message) {
	// Parse request
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(ctx, msg.ID, -32700, "Parse error")
		return
	}

	// Find the prompt handler
	s.mu.RLock()
	handler, exists := s.promptHandlers[params.Name]
	s.mu.RUnlock()

	if !exists {
		s.sendError(ctx, msg.ID, -32602, "Prompt not found")
		return
	}

	// Execute the prompt handler
	messages, err := handler(ctx, params.Arguments)
	if err != nil {
		s.sendError(ctx, msg.ID, -32603, err.Error())
		return
	}

	// Return the prompt result
	result := struct {
		Messages []PromptMessage `json:"messages"`
	}{
		Messages: messages,
	}

	s.sendResult(ctx, msg.ID, result)
}

// NotifyPromptsChanged sends a notification that the prompts list has changed
func (s *Server) NotifyPromptsChanged(ctx context.Context) error {
	notification := &Message{
		JSONRPC: "2.0",
		Method:  "notifications/prompts/list_changed",
	}

	return s.transport.Send(ctx, notification)
}
