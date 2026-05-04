package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	SMTP     SMTPConfig
	AWS      AWSConfig
	Resend   ResendConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         string
	PortPhishing string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

type AWSConfig struct {
	Enabled      bool
	Region       string
	AccessKeyID  string
	SecretKey    string
	SESFromEmail string
	Profile      string
}

type ResendConfig struct {
	Enabled   bool
	APIKey    string
	FromEmail string
}

type AppConfig struct {
	Environment     string
	Debug           bool
	URL             string
	PhishingDomain  string
	PhishingURL     string
	UseCustomDomain bool
	TokenExpireHour int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			PortPhishing: getEnv("PHISHING_PORT", "8081"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "redhook"),
			Password: getEnv("DB_PASSWORD", "redhook"),
			DBName:   getEnv("DB_NAME", "redhook"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpireHour: getEnvAsInt("JWT_EXPIRE_HOUR", 24),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     getEnvAsInt("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", "noreply@redhook.local"),
			UseTLS:   getEnvAsBool("SMTP_USE_TLS", true),
		},
		AWS: AWSConfig{
			Enabled:      getEnvAsBool("AWS_ENABLED", false),
			Region:       getEnv("AWS_REGION", "us-east-1"),
			AccessKeyID:  getEnv("AWS_ACCESS_KEY_ID", ""),
			SecretKey:    getEnv("AWS_SECRET_ACCESS_KEY", ""),
			SESFromEmail: getEnv("AWS_SES_FROM_EMAIL", ""),
			Profile:      getEnv("AWS_PROFILE", "default"),
		},
		Resend: ResendConfig{
			Enabled:   getEnvAsBool("RESEND_ENABLED", false),
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromEmail: getEnv("RESEND_FROM_EMAIL", "RedHook <onboarding@resend.dev>"),
		},
		App: AppConfig{
			Environment:     getEnv("APP_ENV", "development"),
			Debug:           getEnvAsBool("APP_DEBUG", true),
			URL:             getEnv("APP_URL", "http://localhost:8080"),
			PhishingDomain:  getEnv("PHISHING_DOMAIN", "localhost:8081"),
			PhishingURL:     getEnv("PHISHING_URL", "http://localhost:8081"),
			UseCustomDomain: getEnvAsBool("USE_CUSTOM_DOMAIN", false),
			TokenExpireHour: getEnvAsInt("TOKEN_EXPIRE_HOUR", 72),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
