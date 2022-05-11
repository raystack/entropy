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

type Output map[string]interface{}
