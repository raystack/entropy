package log

import (
	"encoding/json"
	"fmt"
	"github.com/odpf/entropy/domain"
)

type Module struct{}

func (m *Module) ID() string {
	return "log"
}

func (m *Module) Apply(r *domain.Resource) (domain.ResourceStatus, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return domain.ResourceStatusError, err
	}
	fmt.Println("=======================================================")
	fmt.Println(string(bytes))
	fmt.Println("=======================================================")
	return domain.ResourceStatusCompleted, nil
}
