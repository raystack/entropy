package model

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func NewError(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func NewErrorf(code int, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func (e Error) String() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *Error) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"code\": %d, \"message\": \"%s\"}", e.Code, e.Message)), nil
}
func (e *Error) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if code, ok := m["code"]; ok {
		if c, ok := code.(float64); ok {
			e.Code = int(c)
		}
	}
	if message, ok := m["message"]; ok {
		if m, ok := message.(string); ok {
			e.Message = m
		}
	}
	return nil
}
