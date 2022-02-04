package metric

import (
	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelic contains the New Relic go-agent configuration
type NewRelicConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"false"`
	AppName string `mapstructure:"appname" default:"entropy-dev"`
	License string `mapstructure:"license"`
}

// RunServer runs the application server
func New(c *NewRelicConfig) (*newrelic.Application, error) {
	return newrelic.NewApplication(
		newrelic.ConfigAppName(c.AppName),
		newrelic.ConfigEnabled(c.Enabled),
		newrelic.ConfigLicense(c.License),
	)
}
