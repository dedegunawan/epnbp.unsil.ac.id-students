package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type OIDCConfig struct {
	OIDCIssuer         string
	OIDCClientID       string
	OIDCClientSecret   string
	OIDCRedirectURI    string
	OIDCLogoutRedirect string
	OIDCLogoutEndpoint string
}

type RedisConfig struct {
	RedisAddress  string
	RedisUsername string
	RedisPassword string
	RedisDB       int
}

type MysqlConfig struct {
	DBHost   string
	DBPort   string
	DBUser   string
	DBPass   string
	DBName   string
	DBParams string
}

func (config MysqlConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		config.DBUser, config.DBPass, config.DBHost, config.DBPort, config.DBName, config.DBParams)
}

type Config struct {
	AppName  string
	AppEnv   string
	HTTPAddr string

	DB1 MysqlConfig

	DBPNBP MysqlConfig

	RedisConfig *RedisConfig

	LogLevel string

	// jwt config
	JWTSecret         string
	JWTIssuer         string
	JWTExpiresMinutes int

	OIDCConfig OIDCConfig
}

func LoadDotEnv() error {
	return godotenv.Load()
}

func FromEnv() Config {
	return Config{
		AppName:  get("APP_NAME", "golang-clean-architecture"),
		AppEnv:   get("APP_ENV", "development"),
		HTTPAddr: get("HTTP_ADDR", ":8080"),

		DB1: MysqlConfig{
			DBHost:   get("DB_HOST", "127.0.0.1"),
			DBPort:   get("DB_PORT", "3306"),
			DBUser:   get("DB_USER", "root"),
			DBPass:   get("DB_PASS", ""),
			DBName:   get("DB_NAME", "yourapp"),
			DBParams: get("DB_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local"),
		},

		DBPNBP: MysqlConfig{
			DBHost:   get("DB_PNBP_HOST", "127.0.0.1"),
			DBPort:   get("DB_PNBP_PORT", "3306"),
			DBUser:   get("DB_PNBP_USER", "root"),
			DBPass:   get("DB_PNBP_PASS", ""),
			DBName:   get("DB_PNBP_NAME", "yourapp"),
			DBParams: get("DB_PNBP_PARAMS", "charset=utf8mb4&parseTime=True&loc=Local"),
		},

		LogLevel: get("LOG_LEVEL", "info"),

		JWTSecret:         get("JWT_SECRET", "please-change-me-32chars-min"),
		JWTIssuer:         get("JWT_ISSUER", "golang-clean-architecture"),
		JWTExpiresMinutes: atoi(get("JWT_EXPIRES_MINUTES", "60")),

		OIDCConfig: OIDCConfig{
			OIDCIssuer:         get("OIDC_ISSUER", "http://localhost:8080/realms/myrealm"),
			OIDCClientID:       get("OIDC_CLIENT_ID", "your-client-id"),
			OIDCClientSecret:   get("OIDC_CLIENT_SECRET", "your-client-secret"),
			OIDCRedirectURI:    get("OIDC_REDIRECT_URI", "http://localhost:8080/auth/callback"),
			OIDCLogoutRedirect: get("OIDC_LOGOUT_REDIRECT", "http://localhost:8080/"),
			OIDCLogoutEndpoint: get("OIDC_LOGOUT_ENDPOINT", "http://localhost:8080/realms/myrealm/protocol/openid-connect/logout"),
		},

		RedisConfig: &RedisConfig{
			RedisAddress:  get("REDIS_ADDRESS", ""),
			RedisUsername: get("REDIS_USERNAME", ""),
			RedisPassword: get("REDIS_PASSWORD", ""),
			RedisDB:       atoi(get("REDIS_DB", "0")),
		},
	}
}
func get(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
