package config

import (
	"github.com/caarlos0/env"
)

type (
	Settings struct {
		DataCentreID    string `env:"DATACENTRE_ID" envDefault:"-"`
		TmpDir          string `env:"TEMP_DIR" envDefault:"tmp"`
		WorkerCount     int    `env:"WORKER_COUNT" envDefault:"8"`
		GoodsBucketSize int    `env:"GOODS_BUCKET_SIZE" envDefault:"50000"`

		MySqlHost          string `env:"MYSQL_HOST" envDefault:"127.0.0.1:3306"`
		MySqlDB            string `env:"MYSQL_DB"`
		MySqlUser          string `env:"MYSQL_USER"`
		MySqlPassword      string `env:"MYSQL_PASSWORD"`
		MySqlMasterReading bool   `env:"MYSQL_MASTERREADING"`

		S3Bucket string `env:"AWS_BUCKET"`

		Bind string `env:"BIND" envDefault:":3001"`
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
