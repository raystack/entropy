package main

import (
	"github.com/odpf/salt/config"
	"github.com/spf13/cobra"

	"github.com/odpf/entropy/internal/server"
	"github.com/odpf/entropy/internal/store/mongodb"
	"github.com/odpf/entropy/pkg/logger"
	"github.com/odpf/entropy/pkg/metric"
)

const configFlag = "config"

// Config contains the application configuration
type Config struct {
	DB       mongodb.DBConfig      `mapstructure:"db"`
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
		opts = append(opts, config.WithPath("./"))
	}

	var cfg Config
	err := config.NewLoader(opts...).Load(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}
