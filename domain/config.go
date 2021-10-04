package domain

// DBConfig contains the database configuration
type DBConfig struct {
	Host string `mapstructure:"host" default:"localhost"`
	Port string `mapstructure:"port" default:"27017"`
	Name string `mapstructure:"name" default:"entropy"`
}

// NewRelic contains the New Relic go-agent configuration
type NewRelicConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"false"`
	AppName string `mapstructure:"appname" default:"entropy-dev"`
	License string `mapstructure:"license"`
}

type LogConfig struct {
	Level string `mapstructure:"level" default:"info"`
}

type ServiceConfig struct {
	Port int    `mapstructure:"port" default:"8080"`
	Host string `mapstructure:"host" default:""`
}

// Config contains the application configuration
type Config struct {
	Service  ServiceConfig  `mapstructure:"service"`
	DB       DBConfig       `mapstructure:"db"`
	NewRelic NewRelicConfig `mapstructure:"newrelic"`
	Log      LogConfig      `mapstructure:"log"`
}
