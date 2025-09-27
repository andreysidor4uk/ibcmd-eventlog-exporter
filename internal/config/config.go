package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	IbcmdPath     string        `yaml:"ibcmd_path"`
	JournalDir    string        `yaml:"journal_dir"`
	LogsDir       string        `yaml:"logs_dir"`
	StartDate     time.Time     `yaml:"start_date"`
	PauseDuration time.Duration `yaml:"pause_duration"`
}

func MustLoad() *Config {
	return mustLoadPath(fetchConfigPath())
}

func mustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	if res == "" {
		res = "config.yml"
	}

	return res
}
