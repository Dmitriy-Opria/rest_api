package config

import (
	"github.com/caarlos0/env"
)

type (
	Settings struct {
		MySqlHost     string `env:"MYSQL_HOST" envDefault:"127.0.0.1:3306"`
		MySqlDB       string `env:"MYSQL_DB" envDefault:"db_test"`
		MySqlUser     string `env:"MYSQL_USER" envDefault:"-"`
		MySqlPassword string `env:"MYSQL_PASSWORD" envDefault:"-"`
		Bind          string `env:"BIND" envDefault:":3001"`
	}
)

var (
	settings = &Settings{}
)

func Parse() {
	env.Parse(settings)
}

func Get() *Settings {
	return settings
}
