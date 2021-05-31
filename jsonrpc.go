package jsonrpc

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	ErrorParseError   = Error{-32700, "Parse error", nil}
	ErrInvalidRequest = Error{-32600, "Invalid Request", nil}
	ErrMethodNotFound = Error{-32601, "Method not found", nil}
	ErrInvalidParams  = Error{-32602, "Invalid params", nil}
	ErrInternalError  = Error{-32603, "Internal error", nil}
	//ErrServerError    = Error{-32000, "Parse error", nil}
)

// TODO: add description
type body struct {
	Version string           `json:"jsonrpc"`
	ID      interface{}      `json:"id"`
	Method  string           `json:"method,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // defined by the server
}

func (e Error) Error() string {
	return fmt.Sprint("jsonrpc: ", strings.ToLower(e.Message))
}
