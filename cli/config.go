package cli

import (
	"fmt"
	"os"

	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
)

const configFlag = "config"

func cmdShowConfigs() *cobra.Command {
	return &cobra.Command{
		Use:   "configs",
		Short: "Display configurations currently loaded",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := loadConfig(cmd)
			if err != nil {
				fmt.Printf("failed to read configs: %v\n", err)
				os.Exit(1)
			}
			_ = yaml.NewEncoder(os.Stdout).Encode(cfg)
		},
	}
}

// Config contains the application configuration
type Config struct {
	DB       mongodb.Config        `mapstructure:"db"`
	Log      logger.LogConfig      `mapstructure:"log"`
	Service  server.Config         `mapstructure:"service"`
	NewRelic metric.NewRelicConfig `mapstructure:"newrelic"`
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
