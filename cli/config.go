package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
)

const configFlag = "config"

// Config contains the application configuration.
type Config struct {
	DB        mongodb.Config        `mapstructure:"db"`
	Log       logger.LogConfig      `mapstructure:"log"`
	Service   server.Config         `mapstructure:"service"`
	NewRelic  metric.NewRelicConfig `mapstructure:"newrelic"`
	PGConnStr string                `mapstructure:"pg_conn_str" default:"postgres://postgres@localhost:5432/entropy?sslmode=disable"`
	Worker    workerConf            `mapstructure:"worker"`
}

type workerConf struct {
	QueueName string `mapstructure:"queue_name"`
	QueueSpec string `mapstructure:"queue_spec"`

	Threads      int           `mapstructure:"threads" default:"1"`
	PollInterval time.Duration `mapstructure:"poll_interval" default:"100ms"`
}

func cmdShowConfigs() *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Display configurations currently loaded",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := loadConfig(cmd)
			if err != nil {
				fmt.Printf("failed to read configs: %v\n", err) // nolint
				os.Exit(1)
			}
			_ = yaml.NewEncoder(os.Stdout).Encode(cfg)
		},
	}
}

func loadConfig(cmd *cobra.Command) (Config, error) {
	var opts []config.LoaderOption

	cfgFile, _ := cmd.Flags().GetString(configFlag)
	if cfgFile != "" {
		opts = append(opts, config.WithFile(cfgFile))
	} else {
		opts = append(opts,
			config.WithPath("./"),
			config.WithName("entropy"),
		)
	}

	var cfg Config
	err := config.NewLoader(opts...).Load(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
