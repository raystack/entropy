package firehose

import "encoding/json"

type Output struct {
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}
