package dbconfig

import (
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Config wraps go-zero SqlConf and lets services explicitly turn read replicas on or off.
type Config struct {
	ReadWriteSplit bool `json:",default=false"`
	DataSource     string
	DriverName     string   `json:",default=mysql"`
	Replicas       []string `json:",optional"`
	Policy         string   `json:",default=round-robin,options=round-robin|random"`
}

func (c Config) ToSqlConf() sqlx.SqlConf {
	conf := sqlx.SqlConf{
		DataSource: c.DataSource,
		DriverName: c.DriverName,
		Policy:     c.Policy,
	}
	if c.ReadWriteSplit {
		conf.Replicas = append([]string(nil), c.Replicas...)
	}

	return conf
}

func (c Config) Validate(name string) error {
	if err := c.ToSqlConf().Validate(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if c.ReadWriteSplit && len(c.Replicas) == 0 {
		return fmt.Errorf("%s: read-write split is enabled but replicas are empty", name)
	}

	return nil
}

func (c Config) MustSqlConf(name string) sqlx.SqlConf {
	if err := c.Validate(name); err != nil {
		panic(err)
	}

	return c.ToSqlConf()
}
