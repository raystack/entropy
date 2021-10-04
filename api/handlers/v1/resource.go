package handlersv1

import (
	"github.com/odpf/entropy/service"
)

type APIServer struct {
	container *service.Container
}

func NewApiServer(container *service.Container) *APIServer {
	return &APIServer{
		container: container,
	}
}
