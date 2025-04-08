package mcp

import (
	"context"
	"encoding/json"
	"net/url"
)

// MCPServer provides a high-level API for creating MCP servers
type MCPServer struct {
	server *Server
}

// NewMCPServer creates a new MCP server
func NewMCPServer(name, version string) *MCPServer {
	return &MCPServer{
		server: NewServer(name, version),
	}
}

// Resource adds a static resource to the server
func (s *MCPServer) Resource(name, uri, description, mimeType string, handler func(ctx context.Context) (string, error)) {
	s.server.AddResource(uri, name, description, mimeType, func(ctx context.Context, uri *url.URL) (ResourceContent, error) {
		text, err := handler(ctx)
		if err != nil {
			return ResourceContent{}, err
		}

		return ResourceContent{
			URI:      uri.String(),
			Text:     text,
			MIMEType: mimeType,
		}, nil
	})
}

// ResourceTemplate adds a dynamic resource template to the server
func (s *MCPServer) ResourceTemplate(name, uriTemplate, description, mimeType string, handler func(ctx context.Context, params map[string]string) (string, error)) {
	template, err := NewResourceTemplate(uriTemplate, description, mimeType)
	if err != nil {
		// Log the error and skip this resource
		// TODO: Better error handling
		return
	}

	s.server.AddResourceTemplate(template, name, func(ctx context.Context, uri *url.URL, params map[string]string) (ResourceContent, error) {
		text, err := handler(ctx, params)
		if err != nil {
			return ResourceContent{}, err
		}

		return ResourceContent{
			URI:      uri.String(),
			Text:     text,
			MIMEType: mimeType,
		}, nil
	})
}

// Tool adds a tool to the server
func (s *MCPServer) Tool(name, description string, schema json.RawMessage, handler func(ctx context.Context, args map[string]interface{}) (string, error)) {
	s.server.AddTool(name, description, schema, func(ctx context.Context, args map[string]interface{}) ([]ToolContent, error) {
		text, err := handler(ctx, args)
		if err != nil {
			return nil, err
		}

		return []ToolContent{{
			Type: "text",
			Text: text,
		}}, nil
	})
}

// Prompt adds a prompt template to the server
func (s *MCPServer) Prompt(name, description string, arguments []PromptArgument, handler func(ctx context.Context, args map[string]interface{}) ([]PromptMessage, error)) {
	s.server.AddPrompt(name, description, arguments, handler)
}

// ConnectStdio connects the server using standard I/O
func (s *MCPServer) ConnectStdio(ctx context.Context) error {
	return s.server.Connect(ctx, NewStdioTransport())
}

// Close terminates the server
func (s *MCPServer) Close() error {
	return s.server.Close()
}

// SendLogMessage sends a logging message notification to the client
func (s *MCPServer) SendLogMessage(ctx context.Context, level string, data interface{}, logger string) error {
	return s.server.SendLogMessage(ctx, level, data, logger)
}

// Helper methods for common log levels
func (s *MCPServer) LogDebug(ctx context.Context, data interface{}, logger string) error {
	return s.server.SendLogMessage(ctx, LogLevelDebug, data, logger)
}

func (s *MCPServer) LogInfo(ctx context.Context, data interface{}, logger string) error {
	return s.server.SendLogMessage(ctx, LogLevelInfo, data, logger)
}

func (s *MCPServer) LogWarning(ctx context.Context, data interface{}, logger string) error {
	return s.server.SendLogMessage(ctx, LogLevelWarning, data, logger)
}

func (s *MCPServer) LogError(ctx context.Context, data interface{}, logger string) error {
	return s.server.SendLogMessage(ctx, LogLevelError, data, logger)
}
