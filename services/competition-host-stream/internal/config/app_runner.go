package config

import "os"

// AppRunnerEnv exports minimal environment variables required when deploying via AWS App Runner.
func AppRunnerEnv(c Config) map[string]string {
	return map[string]string{
		"POSTGRES_DSN": c.PostgresDSN,
		"REDIS_ADDR":   c.RedisAddr,
		"SERVICE_PORT": os.Getenv("SERVICE_PORT"),
		"APP_ENV":      c.AppEnv,
	}
}
