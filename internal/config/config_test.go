package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Test Configuration Loading
func TestConfigLoading(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
		wantErr  bool
	}{
		{
			name: "Valid Configuration",
			envVars: map[string]string{
				"GITEA_URL":            "http://gitea:3000",
				"GITEA_TOKEN":          "test-token",
				"GITEA_WEBHOOK_SECRET": "secret",
			},
			expected: Config{
				ServerAddress:      ":3000",
				LogLevel:           4,
				MaxParallelJudges:  5,
				TestPath:           "test_cases",
				GiteaURL:           "http://gitea:3000",
				GiteaToken:         "test-token",
				GiteaWebhookSecret: "secret",
			},
			wantErr: false,
		},
		{
			name: "Missing Required Fields",
			envVars: map[string]string{
				"SERVER_ADDRESS": ":3000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			cfg, err := Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.GiteaURL, cfg.GiteaURL)
			assert.Equal(t, tt.expected.GiteaToken, cfg.GiteaToken)
			assert.Equal(t, tt.expected.GiteaWebhookSecret, cfg.GiteaWebhookSecret)
		})
	}
}
