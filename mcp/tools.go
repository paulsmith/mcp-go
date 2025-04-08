package mcp

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolHandler is a function that handles tool call requests
type ToolHandler func(ctx context.Context, args map[string]interface{}) ([]ToolContent, error)

// AddTool registers a tool with the server
func (s *Server) AddTool(name, description string, inputSchema json.RawMessage, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tool := Tool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
	}

	// Register the tool
	s.tools = append(s.tools, tool)
	s.toolHandlers[name] = handler
}

// handleListTools handles a tools/list request
func (s *Server) handleListTools(ctx context.Context, msg *Message) {
	s.mu.RLock()
	tools := make([]Tool, len(s.tools))
	copy(tools, s.tools)
	s.mu.RUnlock()

	// Build response
	result := struct {
		Tools []Tool `json:"tools"`
	}{
		Tools: tools,
	}

	s.sendResult(ctx, msg.ID, result)
}

// handleCallTool handles a tools/call request
func (s *Server) handleCallTool(ctx context.Context, msg *Message) {
	// Parse request
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(ctx, msg.ID, -32700, "Parse error")
		return
	}

	// Find the tool handler
	s.mu.RLock()
	handler, exists := s.toolHandlers[params.Name]
	s.mu.RUnlock()

	if !exists {
		s.sendError(ctx, msg.ID, -32602, "Tool not found")
		return
	}

	// Execute the tool
	content, err := handler(ctx, params.Arguments)
	if err != nil {
		// Return the error as a tool result with isError flag
		result := struct {
			Content []ToolContent `json:"content"`
			IsError bool          `json:"isError"`
		}{
			Content: []ToolContent{{
				Type: "text",
				Text: fmt.Sprintf("Error: %v", err),
			}},
			IsError: true,
		}

		s.sendResult(ctx, msg.ID, result)
		return
	}

	// Return the tool result
	result := struct {
		Content []ToolContent `json:"content"`
		IsError bool          `json:"isError"`
	}{
		Content: content,
		IsError: false,
	}

	s.sendResult(ctx, msg.ID, result)
}

// NotifyToolsChanged sends a notification that the tools list has changed
func (s *Server) NotifyToolsChanged(ctx context.Context) error {
	notification := &Message{
		JSONRPC: "2.0",
		Method:  "notifications/tools/list_changed",
	}

	return s.transport.Send(ctx, notification)
}
