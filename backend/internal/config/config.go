package config

import (
	"sync"

	"github.com/spf13/viper"
)

var (
	cfg  *Config
	once sync.Once
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Logging  LoggingConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type ServerConfig struct {
	GRPC GRPCConfig
	HTTP HTTPConfig
}

type GRPCConfig struct {
	Host string
	Port int
}

type HTTPConfig struct {
	Host string
	Port int
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type LoggingConfig struct {
	Level  string
	Format string
}

func Init(configPath string) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 环境变量覆盖
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic("Failed to read config file: " + err.Error())
	}

	viper.Unmarshal(&cfg)
}

func Get() *Config {
	return cfg
}
