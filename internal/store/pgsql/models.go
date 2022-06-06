package pgsql

import "time"

type resourceRecord struct {
	ID              int64     `db:"id"`
	URN             string    `db:"urn"`
	Kind            string    `db:"kind"`
	Name            string    `db:"name"`
	Project         string    `db:"project"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
	SpecConfigs     []byte    `db:"spec_configs"`
	StateStatus     string    `db:"state_status"`
	StateOutput     []byte    `db:"state_output"`
	StateModuleData []byte    `db:"state_module_data"`
}
