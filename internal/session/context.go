package session

import (
	"github.com/yousuf/codebraid-mcp/internal/client"
)

// Context represents a session context with its associated resources
type Context struct {
	SessionID string
	ClientBox *client.ClientBox
}

// NewContext creates a new session context
func NewContext(sessionID string, clientBox *client.ClientBox) *Context {
	return &Context{
		SessionID: sessionID,
		ClientBox: clientBox,
	}
}
