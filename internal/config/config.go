package config

import (
	"os"
	"strconv"
)

type Config struct {
	// Server
	HTTPPort     string
	MTProtoPort  string
	WSPort       string
	GinMode      string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret string

	// Upload
	UploadPath string
	BaseURL    string

	// TURN
	TURNSecret string
	TURNHost   string

	// API Credentials
	APIId   int
	APIHash string
}

func Load() *Config {
	return &Config{
		// Server
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		MTProtoPort: getEnv("MTPROTO_PORT", "8443"),
		WSPort:      getEnv("WS_PORT", "8081"),
		GinMode:     getEnv("GIN_MODE", "release"),

		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "feiji_user"),
		DBPassword: getEnv("DB_PASSWORD", "feiji_pass_2024"),
		DBName:     getEnv("DB_NAME", "feiji_im"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// JWT
		JWTSecret: getEnv("JWT_SECRET", "feiji-jwt-secret-2024"),

		// Upload
		UploadPath: getEnv("UPLOAD_PATH", "/var/www/uploads"),
		BaseURL:    getEnv("BASE_URL", "https://api.zhihang.icu"),

		// TURN
		TURNSecret: getEnv("TURN_SECRET", "feiji-turn-secret-2024"),
		TURNHost:   getEnv("TURN_HOST", "api.zhihang.icu"),

		// API Credentials
		APIId:   getEnvInt("API_ID", 2040000),
		APIHash: getEnv("API_HASH", "A3406DE8D171bb422bb6DDF60E3E70A5"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
