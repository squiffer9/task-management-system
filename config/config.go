package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name    string
	Version string
	Env     string
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	HTTP HTTPServerConfig
	GRPC GRPCServerConfig
}

// HTTPServerConfig holds HTTP server configuration
type HTTPServerConfig struct {
	Port int
}

// GRPCServerConfig holds gRPC server configuration
type GRPCServerConfig struct {
	Port int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	MongoDB MongoDBConfig
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI     string
	Name    string
	Timeout time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWT JWTConfig
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config

	// App config
	cfg.App.Name = viper.GetString("app.name")
	cfg.App.Version = viper.GetString("app.version")
	cfg.App.Env = viper.GetString("app.env")

	// Server config
	cfg.Server.HTTP.Port = viper.GetInt("server.http.port")
	cfg.Server.GRPC.Port = viper.GetInt("server.grpc.port")

	// Database config
	cfg.Database.MongoDB.URI = viper.GetString("database.mongodb.uri")
	cfg.Database.MongoDB.Name = viper.GetString("database.mongodb.name")
	cfg.Database.MongoDB.Timeout = time.Duration(viper.GetInt("database.mongodb.timeout")) * time.Second

	// Auth config
	cfg.Auth.JWT.Secret = viper.GetString("auth.jwt.secret")
	cfg.Auth.JWT.Expiry = time.Duration(viper.GetInt("auth.jwt.expiry")) * time.Hour

	return &cfg, nil
}
