package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	JWT      JWTConfig      `yaml:"jwt"`
	Hash     HashConfig     `yaml:"hash"`
	Database DatabaseConfig `yaml:"database"`
	HTTP     HTTPServer     `yaml:"http"`
	Redis    RedisConfig    `yaml:"redis"`
	Cache    CacheConfig    `yaml:"cache"`
	App      AppConfig      `yaml:"app"`
}

type JWTConfig struct {
	SecretKey       string        `env:"JWT_SECRET_KEY" env-required:"true"`
	AccessTTL       time.Duration `yaml:"access_ttl" env-default:"30m"`
	RefreshTTL      time.Duration `yaml:"refresh_ttl" env-default:"7d"`
	WhitelistPrefix string        `yaml:"whitelist_prefix" env-default:"wl:"`
}

type HashConfig struct {
	Memory     uint32 `yaml:"memory" env-default:"32768"`
	Time       uint32 `yaml:"time" env-default:"3"`
	Threads    uint8  `yaml:"threads" env-default:"2"`
	KeyLen     uint32 `yaml:"key_len" env-default:"32"`
	SaltLength uint8  `yaml:"salt_length" env-default:"16"`
}

type HTTPServer struct {
	Addr           string        `yaml:"addr" env-default:":8080"`
	CORSUrl        []string      `yaml:"cors_url" env-default:"http://localhost:8080"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ReadTimeout    time.Duration `yaml:"read_timeout" env-default:"15s"`
	WriteTimeout   time.Duration `yaml:"write_timeout" env-default:"10s"`
	HandlerTimeout time.Duration `yaml:"handler_timeout" env-default:"5s"`
	CookieSecure   bool          `yaml:"cookie_secure" env-default:"true"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host" env-default:"localhost"`
	Name            string        `yaml:"name" env-required:"true"`
	MaxIdleConns    int32         `yaml:"max_idle_conns" env-default:"25"`
	MaxOpenConns    int32         `yaml:"max_open_conns" env-default:"100"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env-default:"1h"`
	Port            int           `yaml:"port" env-default:"5432"`
	Username        string        `env:"DB_USERNAME" env-required:"true"`
	Password        string        `env:"DB_PASSWORD" env-required:"true"`
	ConnTimeout     time.Duration `yaml:"conn_timeout" env-default:"30s"`
}

type RedisConfig struct {
	Addr            string        `yaml:"addr" env-default:"localhost:6379"`
	Password        string        `env:"REDIS_PASSWORD" env-required:"true"`
	Username        string        `env:"REDIS_USERNAME" env-required:"true"`
	DB              int           `yaml:"db" env-default:"0"`
	MinIdleConns    int           `yaml:"min_idle_conns" env-default:"100"`
	PoolSize        int           `yaml:"pool_size" env-default:"100"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"1s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"1s"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"1h"`
	ConnTimeout     time.Duration `yaml:"conn_timeout" env-default:"10s"`
}

type CacheConfig struct {
	CacheTaskTTL    time.Duration `yaml:"cache_task_ttl" env-default:"1h"`
	TasksPrefix     string        `yaml:"tasks_cache_prefix" env-default:"tasks:"`
	UserTasksPrefix string        `yaml:"user_tasks_cache_prefix" env-default:"user:"`
}

type AppConfig struct {
	MaxTasksPerUser int `yaml:"max_tasks_per_user" env-default:"100"`
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed load config: %w", err)
	}

	return &cfg, nil
}
