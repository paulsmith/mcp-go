package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
)

// Server represents an MCP server
type Server struct {
	// Server identity
	info ServerInfo

	// Capabilities
	capabilities map[string]interface{}

	// Resources
	resources                []Resource
	resourceHandlers         map[string]ResourceHandler
	resourceTemplates        map[string]*ResourceTemplate
	resourceTemplateHandlers map[string]ResourceTemplateHandler

	// Tools
	tools        []Tool
	toolHandlers map[string]ToolHandler

	// Prompts
	prompts        []Prompt
	promptHandlers map[string]PromptHandler

	// Transport
	transport Transport

	// State
	initialized atomic.Bool
	nextID      int64
	mu          sync.RWMutex
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		info: ServerInfo{
			Name:    name,
			Version: version,
		},
		capabilities: map[string]interface{}{
			"resources": map[string]interface{}{},
			"tools":     map[string]interface{}{},
			"prompts":   map[string]interface{}{},
			"logging":   map[string]interface{}{},
		},
		resources:                make([]Resource, 0),
		resourceHandlers:         make(map[string]ResourceHandler),
		resourceTemplates:        make(map[string]*ResourceTemplate),
		resourceTemplateHandlers: make(map[string]ResourceTemplateHandler),
		tools:                    make([]Tool, 0),
		toolHandlers:             make(map[string]ToolHandler),
		prompts:                  make([]Prompt, 0),
		promptHandlers:           make(map[string]PromptHandler),
	}
}

// Connect attaches a transport to the server
func (s *Server) Connect(ctx context.Context, transport Transport) error {
	s.transport = transport

	// Start the message handler
	go s.handleMessages(ctx)

	return nil
}

// Close terminates the server connection
func (s *Server) Close() error {
	if s.transport != nil {
		return s.transport.Close()
	}
	return nil
}

// handleMessages processes incoming messages
func (s *Server) handleMessages(ctx context.Context) {
	for {
		msg, err := s.transport.Receive(ctx)
		if err != nil {
			// Handle error or return if context is done
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			// TODO: Log error
			continue
		}

		go s.handleMessage(ctx, msg)
	}
}

// handleMessage processes a single message
func (s *Server) handleMessage(ctx context.Context, msg *Message) {
	// Skip if no message method is provided
	if msg.Method == "" {
		return
	}

	// Before initialization, only handle initialize messages
	if !s.initialized.Load() && msg.Method != "initialize" {
		s.sendError(ctx, msg.ID, -32002, "Server not initialized")
		return
	}

	// Handle message based on method
	switch msg.Method {
	case "initialize":
		s.handleInitialize(ctx, msg)
	case "initialized":
		// No response needed for this notification
	case "resources/list":
		s.handleListResources(ctx, msg)
	case "resources/read":
		s.handleReadResource(ctx, msg)
	case "tools/list":
		s.handleListTools(ctx, msg)
	case "tools/call":
		s.handleCallTool(ctx, msg)
	case "prompts/list":
		s.handleListPrompts(ctx, msg)
	case "prompts/get":
		s.handleGetPrompt(ctx, msg)
	default:
		s.sendError(ctx, msg.ID, -32601, "Method not found")
	}
}

// handleInitialize processes an initialize request
func (s *Server) handleInitialize(ctx context.Context, msg *Message) {
	// Parse request
	var params struct {
		ProtocolVersion string `json:"protocolVersion"`
		ClientInfo      struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"clientInfo"`
		Capabilities map[string]interface{} `json:"capabilities"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(ctx, msg.ID, -32700, "Parse error")
		return
	}

	// Prepare response
	result := struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		ServerInfo      ServerInfo             `json:"serverInfo"`
		Capabilities    map[string]interface{} `json:"capabilities"`
	}{
		ProtocolVersion: ProtocolVersion,
		ServerInfo:      s.info,
		Capabilities:    s.capabilities,
	}

	// Set server as initialized
	s.initialized.Store(true)

	// Send response
	s.sendResult(ctx, msg.ID, result)
}

// Send a result response
func (s *Server) sendResult(ctx context.Context, id json.RawMessage, result interface{}) {
	if id == nil {
		return // Skip responses to notifications
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		s.sendError(ctx, id, -32603, "Internal error")
		return
	}

	response := &Message{
		ID:      id,
		JSONRPC: "2.0",
		Result:  resultBytes,
	}

	if err := s.transport.Send(ctx, response); err != nil {
		// TODO: Log error
	}
}

// Send an error response
func (s *Server) sendError(ctx context.Context, id json.RawMessage, code int, message string) {
	if id == nil {
		return // Skip responses to notifications
	}

	response := &Message{
		ID:      id,
		JSONRPC: "2.0",
		Error: &ErrorMessage{
			Code:    code,
			Message: message,
		},
	}

	if err := s.transport.Send(ctx, response); err != nil {
		// TODO: Log error
	}
}

// SendLogMessage sends a logging message notification to the client
func (s *Server) SendLogMessage(ctx context.Context, level string, data interface{}, logger string) error {
	params := LoggingMessageParams{
		Level:  level,
		Data:   data,
		Logger: logger,
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	notification := &Message{
		JSONRPC: "2.0",
		Method:  "notifications/message",
		Params:  paramsBytes,
	}

	return s.transport.Send(ctx, notification)
}

// Helper methods for common log levels
func (s *Server) LogDebug(ctx context.Context, data interface{}, logger string) error {
	return s.SendLogMessage(ctx, LogLevelDebug, data, logger)
}

func (s *Server) LogInfo(ctx context.Context, data interface{}, logger string) error {
	return s.SendLogMessage(ctx, LogLevelInfo, data, logger)
}

func (s *Server) LogWarning(ctx context.Context, data interface{}, logger string) error {
	return s.SendLogMessage(ctx, LogLevelWarning, data, logger)
}

func (s *Server) LogError(ctx context.Context, data interface{}, logger string) error {
	return s.SendLogMessage(ctx, LogLevelError, data, logger)
}
