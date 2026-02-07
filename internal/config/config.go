package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	FFmpeg   FFmpegConfig   `mapstructure:"ffmpeg"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Queue    QueueConfig    `mapstructure:"queue"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	MaxConns int    `mapstructure:"max_conns"`
}

type StorageConfig struct {
	Type      string `mapstructure:"type"`
	LocalPath string `mapstructure:"local_path"`
	S3Bucket  string `mapstructure:"s3_bucket"`
	S3Region  string `mapstructure:"s3_region"`
	S3Prefix  string `mapstructure:"s3_prefix"`
}

type FFmpegConfig struct {
	BinaryPath string `mapstructure:"binary_path"`
	Parameters string `mapstructure:"parameters"`
	WorkerCount int   `mapstructure:"worker_count"`
}

type RedisConfig struct {
	SessionAddr string `mapstructure:"session_addr"`
	CacheAddr   string `mapstructure:"cache_addr"`
	Password    string `mapstructure:"password"`
	DB          int    `mapstructure:"db"`
}

type QueueConfig struct {
	Type       string `mapstructure:"type"`
	RabbitMQURL string `mapstructure:"rabbitmq_url"`
	SQSRegion  string `mapstructure:"sqs_region"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("OPENWAN")
	
	// Set defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}
