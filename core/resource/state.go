package resource

import "encoding/json"

const (
	StatusUnspecified Status = "STATUS_UNSPECIFIED" // unknown
	StatusPending     Status = "STATUS_PENDING"     // intermediate
	StatusError       Status = "STATUS_ERROR"       // terminal
	StatusDeleted     Status = "STATUS_DELETED"     // terminal
	StatusCompleted   Status = "STATUS_COMPLETED"   // terminal
)

type State struct {
	Status     Status          `bson:"status"`
	Output     Output          `bson:"output"`
	ModuleData json.RawMessage `bson:"module_data"`
}

type Output map[string]interface{}

type Status string
