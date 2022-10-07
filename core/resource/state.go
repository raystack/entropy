package resource

import (
	"encoding/json"
)

const (
	StatusUnspecified = "STATUS_UNSPECIFIED" // unknown
	StatusPending     = "STATUS_PENDING"     // intermediate
	StatusError       = "STATUS_ERROR"       // terminal
	StatusDeleted     = "STATUS_DELETED"     // terminal
	StatusCompleted   = "STATUS_COMPLETED"   // terminal
)

type State struct {
	Status     string          `json:"status"`
	Output     json.RawMessage `json:"output"`
	ModuleData json.RawMessage `json:"module_data,omitempty"`
	LogOptions LogOptions      `json:"log_options,omitempty"`
}

type LogOptions struct {
	Filters map[string][]string `json:"string,omitempty"`
}

type Output map[string]interface{}

// IsTerminal returns true if state is terminal. A terminal state is
// one where resource needs no further sync.
func (s State) IsTerminal() bool {
	return s.Status == StatusCompleted || s.Status == StatusError
}

// InDeletion returns true if the state represents a resource that is
// scheduled for deletion.
func (s State) InDeletion() bool { return s.Status == StatusDeleted }

func (s State) Clone() State {
	output := make([]byte, len(s.Output))
	copy(output, s.Output)

	newState := State{
		Status:     s.Status,
		Output:     output,
		ModuleData: make([]byte, len(s.ModuleData)),
	}
	copy(newState.ModuleData, s.ModuleData)
	copy(newState.Output, s.Output)
	return newState
}
