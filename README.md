# Model Context Protocol (MCP) Go SDK Documentation

The Model Context Protocol Go SDK is a comprehensive library for building MCP
servers that can expose resources, tools, and prompts to MCP clients like
Claude for Desktop and other MCP-compatible applications.

## Overview

The Go MCP SDK implements the Model Context Protocol specification, allowing
developers to create servers that expose functionality to AI models in a
standardized way. The SDK handles all protocol-level communication, allowing
you to focus on implementing your server's core capabilities.

## Installation

```bash
go get github.com/paulsmith/mcp-go@latest
```

## Usage

```go
import "github.com/paulsmith/mcp-go/mcp"
```

## Core Components

The SDK is built around several key components:

### MCPServer

The `MCPServer` struct is the main entry point for creating MCP servers. It
provides a simplified high-level API for common operations.

```go
// Create a new MCP server
server := mcp.NewMCPServer("MyServer", "1.0.0")
```

### Server

The `Server` struct is the underlying implementation that handles protocol
details. Most users should use `MCPServer` instead, but `Server` is available
for advanced use cases.

### Transport

The `Transport` interface defines how messages are exchanged between the client
and server. The SDK includes a `StdioTransport` implementation for standard I/O
communication.

## Key Features

### Resources

Resources allow your server to expose data and content to LLMs:

```go
// Add a static resource
server.Resource("Example Resource", "example://resource", "Example resource", "text/plain", 
    func(ctx context.Context) (string, error) {
        return "This is an example resource", nil
    })

// Add a dynamic resource template
server.ResourceTemplate("User Data", "user://{userId}", "User data for a specific user", "application/json",
    func(ctx context.Context, params map[string]string) (string, error) {
        userId := params["userId"]
        // Fetch user data based on userId
        return fmt.Sprintf(`{"id": "%s", "name": "Example User"}`, userId), nil
    })
```

### Tools

Tools are functions that can be called by LLMs to perform actions:

```go
// Add a tool
calculatorSchema := json.RawMessage(`{
    "type": "object",
    "properties": {
        "operation": {
            "type": "string",
            "enum": ["add", "subtract", "multiply", "divide"]
        },
        "a": {"type": "number"},
        "b": {"type": "number"}
    },
    "required": ["operation", "a", "b"]
}`)

server.Tool("calculate", "Perform a calculation", calculatorSchema,
    func(ctx context.Context, args map[string]interface{}) (string, error) {
        operation := args["operation"].(string)
        a := args["a"].(float64)
        b := args["b"].(float64)
        
        var result float64
        switch operation {
        case "add":
            result = a + b
        case "subtract":
            result = a - b
        case "multiply":
            result = a * b
        case "divide":
            if b == 0 {
                return "", fmt.Errorf("division by zero")
            }
            result = a / b
        default:
            return "", fmt.Errorf("unknown operation: %s", operation)
        }
        
        return fmt.Sprintf("Result: %g", result), nil
    })
```

### Prompts

Prompts are reusable templates that guide LLM interactions:

```go
// Add a prompt
server.Prompt("greeting", "Generate a personalized greeting", 
    []mcp.PromptArgument{
        {Name: "name", Description: "Name of the person to greet", Required: true},
        {Name: "language", Description: "Language to use (en, es, fr)", Required: false},
    },
    func(ctx context.Context, args map[string]interface{}) ([]mcp.PromptMessage, error) {
        name := args["name"].(string)
        language := "en"
        if lang, ok := args["language"].(string); ok {
            language = lang
        }
        
        var greeting string
        switch language {
        case "en":
            greeting = fmt.Sprintf("Hello, %s!", name)
        case "es":
            greeting = fmt.Sprintf("¡Hola, %s!", name)
        case "fr":
            greeting = fmt.Sprintf("Bonjour, %s!", name)
        default:
            greeting = fmt.Sprintf("Hello, %s!", name)
        }
        
        content, _ := json.Marshal(mcp.TextContent{
            Type: "text",
            Text: greeting,
        })
        
        return []mcp.PromptMessage{
            {Role: "user", Content: content},
        }, nil
    })
```

### Logging

The SDK includes built-in support for sending logs to clients:

```go
// Send log messages with different levels
server.LogDebug(ctx, "Debug message", "example-logger")
server.LogInfo(ctx, "Info message", "example-logger")
server.LogWarning(ctx, "Warning message", "example-logger")
server.LogError(ctx, "Error message", "example-logger")

// Or use the generic method
server.SendLogMessage(ctx, mcp.LogLevelNotice, "Notice message", "example-logger")
```

## Complete Example

Here's a complete example of a simple calculator server:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/paulsmith/mcp-go/mcp"
)

func main() {
    // Create a new MCP server
    server := mcp.NewMCPServer("Calculator", "1.0.0")

    // Define the calculator schema
    calculatorSchema := json.RawMessage(`{
        "type": "object",
        "properties": {
            "operation": {
                "type": "string",
                "enum": ["add", "subtract", "multiply", "divide"],
                "description": "Operation to perform"
            },
            "a": {
                "type": "number",
                "description": "First operand"
            },
            "b": {
                "type": "number",
                "description": "Second operand"
            }
        },
        "required": ["operation", "a", "b"]
    }`)

    // Add a calculator tool
    server.Tool("calculate", "Perform a calculation", calculatorSchema,
        func(ctx context.Context, args map[string]interface{}) (string, error) {
            operation := args["operation"].(string)
            a := args["a"].(float64)
            b := args["b"].(float64)

            var result float64
            switch operation {
            case "add":
                result = a + b
            case "subtract":
                result = a - b
            case "multiply":
                result = a * b
            case "divide":
                if b == 0 {
                    return "", fmt.Errorf("division by zero")
                }
                result = a / b
            default:
                return "", fmt.Errorf("unknown operation: %s", operation)
            }

            return fmt.Sprintf("Result: %g", result), nil
        })

    // Connect using standard I/O
    if err := server.ConnectStdio(context.Background()); err != nil {
        fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
        os.Exit(1)
    }

    // Log startup
    server.LogInfo(context.Background(), "Calculator server started", "main")

    // Run until context is cancelled
    select {}
}
```

## API Reference

### MCPServer

```go
type MCPServer struct {
    server *Server
}

// NewMCPServer creates a new MCP server
func NewMCPServer(name, version string) *MCPServer

// Resource adds a static resource to the server
func (s *MCPServer) Resource(name, uri, description, mimeType string, handler func(ctx context.Context) (string, error))

// ResourceTemplate adds a dynamic resource template to the server
func (s *MCPServer) ResourceTemplate(name, uriTemplate, description, mimeType string, handler func(ctx context.Context, params map[string]string) (string, error))

// Tool adds a tool to the server
func (s *MCPServer) Tool(name, description string, schema json.RawMessage, handler func(ctx context.Context, args map[string]interface{}) (string, error))

// Prompt adds a prompt template to the server
func (s *MCPServer) Prompt(name, description string, arguments []PromptArgument, handler func(ctx context.Context, args map[string]interface{}) ([]PromptMessage, error))

// ConnectStdio connects the server using standard I/O
func (s *MCPServer) ConnectStdio(ctx context.Context) error

// Close terminates the server
func (s *MCPServer) Close() error

// SendLogMessage sends a logging message notification to the client
func (s *MCPServer) SendLogMessage(ctx context.Context, level string, data interface{}, logger string) error

// Helper methods for common log levels
func (s *MCPServer) LogDebug(ctx context.Context, data interface{}, logger string) error
func (s *MCPServer) LogInfo(ctx context.Context, data interface{}, logger string) error
func (s *MCPServer) LogWarning(ctx context.Context, data interface{}, logger string) error
func (s *MCPServer) LogError(ctx context.Context, data interface{}, logger string) error
```

### Server

```go
type Server struct {
    // Server identity
    info ServerInfo

    // Capabilities
    capabilities map[string]interface{}

    // Resources, tools, prompts, etc.
    // ...
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server

// Connect attaches a transport to the server
func (s *Server) Connect(ctx context.Context, transport Transport) error

// Close terminates the server connection
func (s *Server) Close() error

// AddResource registers a static resource with the server
func (s *Server) AddResource(uri, name, description, mimeType string, handler ResourceHandler)

// AddResourceTemplate registers a dynamic resource template with the server
func (s *Server) AddResourceTemplate(template *ResourceTemplate, name string, handler ResourceTemplateHandler)

// AddTool registers a tool with the server
func (s *Server) AddTool(name, description string, inputSchema json.RawMessage, handler ToolHandler)

// AddPrompt registers a prompt with the server
func (s *Server) AddPrompt(name, description string, arguments []PromptArgument, handler PromptHandler)

// NotifyResourcesChanged sends a notification that the resources list has changed
func (s *Server) NotifyResourcesChanged(ctx context.Context) error

// NotifyResourceUpdated sends a notification that a resource has been updated
func (s *Server) NotifyResourceUpdated(ctx context.Context, uri string) error

// NotifyToolsChanged sends a notification that the tools list has changed
func (s *Server) NotifyToolsChanged(ctx context.Context) error

// NotifyPromptsChanged sends a notification that the prompts list has changed
func (s *Server) NotifyPromptsChanged(ctx context.Context) error
```

### Transport

```go
type Transport interface {
    // Send transmits a message through the transport
    Send(ctx context.Context, msg *Message) error

    // Receive waits for and returns the next incoming message
    Receive(ctx context.Context) (*Message, error)

    // Close terminates the transport connection
    Close() error
}
```

### StdioTransport

```go
type StdioTransport struct {
    reader    *bufio.Reader
    writer    *bufio.Writer
    writeLock sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport
```

### Resource Types

```go
type Resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MIMEType    string `json:"mimeType,omitempty"`
}

type ResourceContent struct {
    URI      string `json:"uri"`
    Text     string `json:"text,omitempty"`
    Blob     []byte `json:"blob,omitempty"`
    MIMEType string `json:"mimeType,omitempty"`
}

type ResourceHandler func(ctx context.Context, uri *url.URL) (ResourceContent, error)
type ResourceTemplateHandler func(ctx context.Context, uri *url.URL, params map[string]string) (ResourceContent, error)
```

### Tool Types

```go
type Tool struct {
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    InputSchema json.RawMessage `json:"inputSchema"`
}

type ToolContent struct {
    Type string `json:"type"`
    Text string `json:"text,omitempty"`
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) ([]ToolContent, error)
```

### Prompt Types

```go
type Prompt struct {
    Name        string           `json:"name"`
    Description string           `json:"description,omitempty"`
    Arguments   []PromptArgument `json:"arguments,omitempty"`
}

type PromptArgument struct {
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    Required    bool   `json:"required,omitempty"`
}

type PromptMessage struct {
    Role    string          `json:"role"`
    Content json.RawMessage `json:"content"`
}

type PromptHandler func(ctx context.Context, args map[string]interface{}) ([]PromptMessage, error)
```

### Logging Types

```go
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

type LoggingMessageParams struct {
    Level  string      `json:"level"`
    Data   interface{} `json:"data"`
    Logger string      `json:"logger,omitempty"`
}
```

## Best Practices

1. **Error Handling**: Always handle errors properly in your resource, tool,
   and prompt handlers. Return meaningful error messages to help users
   understand issues.

2. **Context Usage**: Respect the context passed to handlers. Check for
   cancellation and propagate it appropriately.

3. **Resource Organization**: Group related resources and provide clear
   descriptions to help users understand your server's capabilities.

4. **Input Validation**: Validate all inputs in your tool handlers to prevent
   unexpected behavior.

5. **Logging**: Use appropriate log levels to provide useful information
   without overwhelming clients.

6. **Security**: Only expose necessary functionality and validate all inputs to
   prevent security issues.

## Troubleshooting

### Server not connecting to client

- Ensure your server is properly configured with the correct name and version.
- Check that you're using the transport compatible with your client (e.g.,
  `StdioTransport` for Claude Desktop).
- Verify your `claude_desktop_config.json` if using Claude Desktop.

### Tools or resources not appearing

- Make sure you're adding tools and resources before connecting the server.
- Check that your tool and resource handlers are properly implemented.
- Verify that JSON schemas for tools are valid.

### Errors during tool execution

- Implement proper error handling in your tool handlers.
- Return clear error messages that explain what went wrong.
- Check input types and validate arguments before using them.

## License

[MIT](COPYING)
