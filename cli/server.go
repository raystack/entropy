package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/odpf/entropy/core"
	handlersv1 "github.com/odpf/entropy/internal/server/v1"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/modules/firehose"
	"github.com/odpf/entropy/modules/kubernetes"
	"github.com/odpf/entropy/pkg/logger"
)

func GetServer(ctx context.Context, cmd *cobra.Command) *handlersv1.APIServer {
	c, err := loadConfig(cmd)
	if err != nil {
		return nil
	}

	_, cancel := context.WithCancel(ctx)
	defer cancel()

	zapLog, err := logger.New(&c.Log)
	if err != nil {
		return nil
	}

	moduleRegistry := setupRegistry(zapLog,
		kubernetes.Module,
		firehose.Module,
	)

	mongoStore, err := mongodb.Connect(c.DB)
	if err != nil {
		return nil
	}
	resourceRepository := mongodb.NewResourceRepository(mongoStore)
	resourceService := core.New(resourceRepository, moduleRegistry, time.Now, zapLog)
	server := handlersv1.NewAPIServer(resourceService)

	return server
}
