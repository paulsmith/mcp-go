package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/paulsmith/mcpgo/mcp"
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

	// Connect to standard I/O
	if err := server.ConnectStdio(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
		os.Exit(1)
	}

	// Run until context is cancelled
	select {}
}

