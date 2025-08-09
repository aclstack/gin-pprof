package http

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/aclstack/gin-pprof/pkg/core"
)

// GinContext adapts gin.Context to core.HTTPContext interface
type GinContext struct {
	ctx *gin.Context
}

// NewGinContext creates a new GinContext
func NewGinContext(ctx *gin.Context) core.HTTPContext {
	return &GinContext{ctx: ctx}
}

// GetPath returns the route template (e.g., "/users/:id")
func (g *GinContext) GetPath() string {
	path := g.ctx.FullPath()
	if path == "" {
		// Fallback to request path if route template is not available
		path = g.ctx.Request.URL.Path
	}
	return path
}

// GetMethod returns the HTTP method
func (g *GinContext) GetMethod() string {
	return g.ctx.Request.Method
}

// GetHeaders returns request headers
func (g *GinContext) GetHeaders() map[string]string {
	headers := make(map[string]string)
	for key, values := range g.ctx.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0] // Take first value
		}
	}
	return headers
}

// SetContext sets a key-value pair in gin context
func (g *GinContext) SetContext(key, value interface{}) {
	g.ctx.Set(keyToString(key), value)
}

// GetContext gets a value from gin context by key
func (g *GinContext) GetContext(key interface{}) interface{} {
	value, exists := g.ctx.Get(keyToString(key))
	if !exists {
		return nil
	}
	return value
}

// GetRequestPath returns the actual request path
func (g *GinContext) GetRequestPath() string {
	return g.ctx.Request.URL.Path
}

// GinPathMatcher implements path matching for Gin routes
type GinPathMatcher struct{}

// NewGinPathMatcher creates a new GinPathMatcher
func NewGinPathMatcher() core.PathMatcher {
	return &GinPathMatcher{}
}

// Match checks if actual path matches the gin route template
func (g *GinPathMatcher) Match(template, actual string) bool {
	templateParts := strings.Split(strings.Trim(template, "/"), "/")
	actualParts := strings.Split(strings.Trim(actual, "/"), "/")

	if len(templateParts) != len(actualParts) {
		return false
	}

	for i, templatePart := range templateParts {
		if strings.HasPrefix(templatePart, ":") || templatePart == "*" {
			// This is a parameter or wildcard, it matches anything
			continue
		}
		if templatePart != actualParts[i] {
			return false
		}
	}

	return true
}

// ExtractParams extracts parameters from actual path using gin route template
func (g *GinPathMatcher) ExtractParams(template, actual string) map[string]string {
	params := make(map[string]string)
	
	templateParts := strings.Split(strings.Trim(template, "/"), "/")
	actualParts := strings.Split(strings.Trim(actual, "/"), "/")

	if len(templateParts) != len(actualParts) {
		return params
	}

	for i, templatePart := range templateParts {
		if strings.HasPrefix(templatePart, ":") {
			// Extract parameter name (remove the ':')
			paramName := templatePart[1:]
			params[paramName] = actualParts[i]
		}
	}

	return params
}

// keyToString converts interface{} key to string for gin context
func keyToString(key interface{}) string {
	if str, ok := key.(string); ok {
		return str
	}
	return ""
}