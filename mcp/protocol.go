// protocol.go
package mcp

import (
	"encoding/json"
	"fmt"
)

// Protocol constants
const (
	ProtocolVersion = "2024-11-05"
)

// Message represents a protocol message
type Message struct {
	ID      json.RawMessage `json:"id,omitempty"`
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorMessage   `json:"error,omitempty"`
}

func (m *Message) GetIDString() string {
	if m.ID == nil {
		return ""
	}

	var strID string
	if err := json.Unmarshal(m.ID, &strID); err == nil {
		return strID
	}

	var numID float64
	if err := json.Unmarshal(m.ID, &numID); err == nil {
		return fmt.Sprintf("%v", numID)
	}

	return string(m.ID)
}

// ErrorMessage represents an error response
type ErrorMessage struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ServerInfo contains information about the server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Resource represents a resource that can be accessed by clients
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MIMEType    string `json:"mimeType,omitempty"`
}

// Tool represents a tool that can be called by clients
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ToolContent represents content returned by a tool
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Prompt represents a prompt template
type Prompt struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptArgument defines an argument to a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// TextContent represents text content
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Add these constants for logging levels
const (
	LogLevelDebug     = "debug"
	LogLevelInfo      = "info"
	LogLevelNotice    = "notice"
	LogLevelWarning   = "warning"
	LogLevelError     = "error"
	LogLevelCritical  = "critical"
	LogLevelAlert     = "alert"
	LogLevelEmergency = "emergency"
)

// LoggingMessageParams represents the parameters for a logging message notification
type LoggingMessageParams struct {
	Level  string      `json:"level"`
	Data   interface{} `json:"data"`
	Logger string      `json:"logger,omitempty"`
}
