package jsonrpc2

import (
	"fmt"
	"strings"
)

var (
	ErrorParseError   = &Error{-32700, "Parse error", nil}
	ErrInvalidRequest = &Error{-32600, "Invalid Request", nil}
	ErrMethodNotFound = &Error{-32601, "Method not found", nil}
	ErrInvalidParams  = &Error{-32602, "Invalid params", nil}
	ErrInternalError  = &Error{-32603, "Internal error", nil}
	//ErrServerError    = Error{-32000, "Parse error", nil}
)

// Error represents a JSON-RPC error, it implements the error interface.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // defined by the server
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprint("jsonrpc: ", strings.ToLower(e.Message))
}
