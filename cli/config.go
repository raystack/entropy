package cli

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/goto/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/goto/entropy/pkg/errors"
	"github.com/goto/entropy/pkg/logger"
	"github.com/goto/entropy/pkg/telemetry"
)

const configFlag = "config"

// Config contains the application configuration.
type Config struct {
	Log       logger.LogConfig `mapstructure:"log"`
	Worker    workerConf       `mapstructure:"worker"`
	Service   serveConfig      `mapstructure:"service"`
	PGConnStr string           `mapstructure:"pg_conn_str" default:"postgres://postgres@localhost:5432/entropy?sslmode=disable"`
	Telemetry telemetry.Config `mapstructure:"telemetry"`
}

type serveConfig struct {
	Host string `mapstructure:"host" default:""`
	Port int    `mapstructure:"port" default:"8080"`

	HTTPAddr string `mapstructure:"http_addr" default:":8081"`
}

type workerConf struct {
	QueueName string `mapstructure:"queue_name" default:"entropy_jobs"`
	QueueSpec string `mapstructure:"queue_spec" default:"postgres://postgres@localhost:5432/entropy?sslmode=disable"`

	Threads      int           `mapstructure:"threads" default:"1"`
	PollInterval time.Duration `mapstructure:"poll_interval" default:"100ms"`
}

func (serveCfg serveConfig) httpAddr() string { return serveCfg.HTTPAddr }

func (serveCfg serveConfig) grpcAddr() string {
	return fmt.Sprintf("%s:%d", serveCfg.Host, serveCfg.Port)
}

func cmdShowConfigs() *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Display configurations currently loaded",
		RunE: handleErr(func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(cmd)
			if err != nil {
				fatalExitf("failed to read configs: %v", err)
			}
			return yaml.NewEncoder(os.Stdout).Encode(cfg)
		}),
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
	if errors.As(err, &config.ConfigFileNotFoundError{}) {
		log.Println(err)
	} else {
		return cfg, err
	}

	return cfg, nil
}
