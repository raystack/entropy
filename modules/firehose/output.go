package firehose

import "encoding/json"

type Output struct {
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"release_name"`
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}
