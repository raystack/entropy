package firehose

import "encoding/json"

type moduleData struct {
	PendingSteps  []string `json:"pending_steps"`
	ResetTo       string   `json:"reset_to,omitempty"`
	StateOverride string   `json:"state_override,omitempty"`
}

func (md moduleData) JSON() json.RawMessage {
	bytes, err := json.Marshal(md)
	if err != nil {
		panic(err)
	}
	return bytes
}
