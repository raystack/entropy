package firehose

import "encoding/json"

type moduleData struct {
	PendingSteps   []string `json:"pending_steps"`
	ResetTimestamp int64    `json:"reset_timestamp,omitempty"`
	StateOverride  string   `json:"state_override,omitempty"`
}

func (md moduleData) JSON() json.RawMessage {
	bytes, err := json.Marshal(md)
	if err != nil {
		panic(err)
	}
	return bytes
}
