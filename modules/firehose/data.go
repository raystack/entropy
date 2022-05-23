package firehose

import "encoding/json"

type moduleData struct {
	PendingSteps     []string `json:"pending_steps"`
	LastReplicaCount int64    `json:"last_replica_count"`
}

func (md moduleData) JSON() json.RawMessage {
	bytes, err := json.Marshal(md)
	if err != nil {
		panic(err)
	}
	return bytes
}
