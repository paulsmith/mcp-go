// stdio.go
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
)

// StdioTransport implements the Transport interface using standard input/output
type StdioTransport struct {
	reader    *bufio.Reader
	writer    *bufio.Writer
	writeLock sync.Mutex
}

// NewStdioTransport creates a new stdio transport
func NewStdioTransport() *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Send transmits a message through the transport
func (t *StdioTransport) Send(ctx context.Context, msg *Message) error {
	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Write message
	if _, err := t.writer.Write(data); err != nil {
		return err
	}

	if _, err := t.writer.Write([]byte("\n")); err != nil {
		return err
	}

	return t.writer.Flush()
}

// Receive waits for and returns the next incoming message
func (t *StdioTransport) Receive(ctx context.Context) (*Message, error) {
	data, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// Close terminates the transport connection
func (t *StdioTransport) Close() error {
	return nil // Nothing to close for stdio
}

