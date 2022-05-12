package resource

import "encoding/json"

const (
	StatusUnspecified = "STATUS_UNSPECIFIED" // unknown
	StatusPending     = "STATUS_PENDING"     // intermediate
	StatusError       = "STATUS_ERROR"       // terminal
	StatusDeleted     = "STATUS_DELETED"     // terminal
	StatusCompleted   = "STATUS_COMPLETED"   // terminal
)

type State struct {
	Status     string          `json:"status" bson:"status"`
	Output     Output          `json:"output" bson:"output"`
	ModuleData json.RawMessage `json:"module_data" bson:"module_data"`
}

// IsTerminal returns true if state is terminal. A terminal state is
// one where resource needs no further sync.
func (s State) IsTerminal() bool {
	return s.Status == StatusCompleted || s.Status == StatusError
}

// InDeletion returns true if the state represents a resource that is
// scheduled for deletion.
func (s State) InDeletion() bool { return s.Status == StatusDeleted }

func (s State) Clone() State {
	newState := State{
		Status:     s.Status,
		Output:     map[string]interface{}{},
		ModuleData: make([]byte, len(s.ModuleData)),
	}
	copy(newState.ModuleData, s.ModuleData)
	for k, v := range s.Output {
		newState.Output[k] = v
	}
	return newState
}

type Output map[string]interface{}
