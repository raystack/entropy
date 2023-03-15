package kubernetes

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/version"

	"github.com/goto/entropy/pkg/kube"
)

type Output struct {
	Configs    kube.Config  `json:"configs"`
	ServerInfo version.Info `json:"server_info"`
}

func (out Output) JSON() []byte {
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	return b
}
