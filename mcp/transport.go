package mcp

import (
	"context"
)

// Transport defines the interface for MCP communication channels
type Transport interface {
	// Send transmits a message through the transport
	Send(ctx context.Context, msg *Message) error

	// Receive waits for and returns the next incoming message
	Receive(ctx context.Context) (*Message, error)

	// Close terminates the transport connection
	Close() error
}

