package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	MQEngine   `yaml:"mq_engine"`
}

type MQEngine struct {
	Queues       []string `yaml:"queues" env-required:"true"`
	MessageLimit int      `yaml:"message_limit" env-default:"100"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad() *Config {
	//os.Setenv("DUMMYMQ_CONFIG_PATH", "/Users/megafon/Documents/go/dummymq/config/local.yaml")
	configPath := os.Getenv("DUMMYMQ_CONFIG_PATH")
	if configPath == "" {
		log.Fatal("DUMMYMQ_CONFIG_PATH is not set")
	}

	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
