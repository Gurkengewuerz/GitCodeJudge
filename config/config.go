package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Server configuration
	ServerAddress     string `envconfig:"SERVER_ADDRESS" default:":3000"`
	LogLevel          int    `envconfig:"LOG_LEVEL" default:"4"`
	MaxParallelJudges int    `envconfig:"MAX_PARALLEL_JUDGES" default:"5"`
	TestPath          string `envconfig:"TESTS_PATH" default:"test_cases"`

	// Database
	DatabasePath string `envconfig:"DB_PATH" default:"database/"`
	DatabaseTTL  int    `envconfig:"DB_TTL" default:"0"`

	// PDF
	PDFFooterCopyright     string `envconfig:"PDF_FOOTER_COPYRIGHT" default:""`
	PDFFooterGeneratedWith string `envconfig:"PDF_FOOTER_GENERATEDWITH" default:"Generated with https://github.com/Gurkengewuerz/GitCodeJudge"`

	// Gitea configuration
	GiteaURL           string `envconfig:"GITEA_URL" required:"true"`
	GiteaToken         string `envconfig:"GITEA_TOKEN" required:"true"`
	GiteaWebhookSecret string `envconfig:"GITEA_WEBHOOK_SECRET" required:"true"`

	// Docker configuration
	DockerImage   string `envconfig:"DOCKER_IMAGE" default:"ghcr.io/gurkengewuerz/gitcodejudge-judge:latest"`
	DockerNetwork string `envconfig:"DOCKER_NETWORK" default:"none"`
	DockerTimeout int    `envconfig:"DOCKER_TIMEOUT" default:"30"`
}

var CFG *Config

func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}
	CFG = cfg
	return cfg, nil
}
