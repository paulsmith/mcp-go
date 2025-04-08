package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ResourceHandler is a function that handles resource read requests for static URIs
type ResourceHandler func(ctx context.Context, uri *url.URL) (ResourceContent, error)

// ResourceTemplateHandler is a function that handles resource read requests for URI templates
type ResourceTemplateHandler func(ctx context.Context, uri *url.URL, params map[string]string) (ResourceContent, error)

// ResourceContent represents content returned by a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	Text     string `json:"text,omitempty"`
	Blob     []byte `json:"blob,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
}

// ResourceTemplate represents a URI template for dynamic resources
type ResourceTemplate struct {
	Template    string
	regex       *regexp.Regexp
	paramNames  []string
	Description string
	MIMEType    string
}

// NewResourceTemplate creates a new resource template
func NewResourceTemplate(template, description, mimeType string) (*ResourceTemplate, error) {
	// Extract parameter names from the template
	paramPattern := regexp.MustCompile(`\{([^{}]+)\}`)
	matches := paramPattern.FindAllStringSubmatch(template, -1)

	paramNames := make([]string, 0, len(matches))
	for _, match := range matches {
		paramNames = append(paramNames, match[1])
	}

	// Convert template to regex for matching
	regexPattern := template
	for _, param := range paramNames {
		regexPattern = strings.Replace(regexPattern, "{"+param+"}", "([^/]+)", 1)
	}
	regexPattern = "^" + regexPattern + "$"

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	return &ResourceTemplate{
		Template:    template,
		regex:       regex,
		paramNames:  paramNames,
		Description: description,
		MIMEType:    mimeType,
	}, nil
}

// Match checks if a URI matches this template and extracts parameters
func (t *ResourceTemplate) Match(uri string) (map[string]string, bool) {
	matches := t.regex.FindStringSubmatch(uri)
	if matches == nil {
		return nil, false
	}

	params := make(map[string]string)
	for i, name := range t.paramNames {
		params[name] = matches[i+1]
	}

	return params, true
}

// AddResource registers a static resource with the server
func (s *Server) AddResource(uri, name, description, mimeType string, handler ResourceHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resource := Resource{
		URI:         uri,
		Name:        name,
		Description: description,
		MIMEType:    mimeType,
	}

	// Register the resource
	s.resources = append(s.resources, resource)
	s.resourceHandlers[uri] = handler
}

// AddResourceTemplate registers a dynamic resource template with the server
func (s *Server) AddResourceTemplate(template *ResourceTemplate, name string, handler ResourceTemplateHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resource := Resource{
		URI:         template.Template, // Store the template pattern
		Name:        name,
		Description: template.Description,
		MIMEType:    template.MIMEType,
	}

	// Register the resource template
	s.resourceTemplates[template.Template] = template
	s.resourceTemplateHandlers[template.Template] = handler
	s.resources = append(s.resources, resource)
}

// handleListResources handles a resources/list request
func (s *Server) handleListResources(ctx context.Context, msg *Message) {
	s.mu.RLock()
	resources := make([]Resource, len(s.resources))
	copy(resources, s.resources)
	s.mu.RUnlock()

	// Build response
	result := struct {
		Resources []Resource `json:"resources"`
	}{
		Resources: resources,
	}

	s.sendResult(ctx, msg.ID, result)
}

// handleReadResource handles a resources/read request
func (s *Server) handleReadResource(ctx context.Context, msg *Message) {
	// Parse request
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(msg.Params, &params); err != nil {
		s.sendError(ctx, msg.ID, -32700, "Parse error")
		return
	}

	// Parse URI
	uri, err := url.Parse(params.URI)
	if err != nil {
		s.sendError(ctx, msg.ID, -32602, "Invalid URI")
		return
	}

	// Try static resources first
	s.mu.RLock()
	handler, exists := s.resourceHandlers[uri.String()]
	s.mu.RUnlock()

	if exists {
		content, err := handler(ctx, uri)
		if err != nil {
			s.sendError(ctx, msg.ID, -32603, fmt.Sprintf("Error reading resource: %v", err))
			return
		}

		result := struct {
			Contents []ResourceContent `json:"contents"`
		}{
			Contents: []ResourceContent{content},
		}

		s.sendResult(ctx, msg.ID, result)
		return
	}

	// Try resource templates
	s.mu.RLock()
	for templateStr, template := range s.resourceTemplates {
		params, matches := template.Match(uri.String())
		if matches {
			handler := s.resourceTemplateHandlers[templateStr]
			s.mu.RUnlock()

			content, err := handler(ctx, uri, params)
			if err != nil {
				s.sendError(ctx, msg.ID, -32603, fmt.Sprintf("Error reading resource: %v", err))
				return
			}

			result := struct {
				Contents []ResourceContent `json:"contents"`
			}{
				Contents: []ResourceContent{content},
			}

			s.sendResult(ctx, msg.ID, result)
			return
		}
	}
	s.mu.RUnlock()

	// Resource not found
	s.sendError(ctx, msg.ID, -32602, "Resource not found")
}

// NotifyResourcesChanged sends a notification that the resources list has changed
func (s *Server) NotifyResourcesChanged(ctx context.Context) error {
	notification := &Message{
		JSONRPC: "2.0",
		Method:  "notifications/resources/list_changed",
	}

	return s.transport.Send(ctx, notification)
}

// NotifyResourceUpdated sends a notification that a resource has been updated
func (s *Server) NotifyResourceUpdated(ctx context.Context, uri string) error {
	params := struct {
		URI string `json:"uri"`
	}{
		URI: uri,
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	notification := &Message{
		JSONRPC: "2.0",
		Method:  "notifications/resources/updated",
		Params:  paramsBytes,
	}

	return s.transport.Send(ctx, notification)
}
