package firehose

import "encoding/json"

type moduleData struct {
	PendingSteps []string `json:"pending_steps"`
}

func (md moduleData) JSON() json.RawMessage {
	bytes, err := json.Marshal(md)
	if err != nil {
		panic(err)
	}
	return bytes
}
