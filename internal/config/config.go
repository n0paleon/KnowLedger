package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Log      LogConfig      `mapstructure:"log"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	GC       GCConfig       `mapstructure:"gc"`
	Admin    AdminConfig    `mapstructure:"admin"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Dev  bool   `mapstructure:"dev"`
}

type LogConfig struct {
	Level  string   `mapstructure:"level"`
	Output []string `mapstructure:"output"`
}

type HTTPConfig struct {
	ListenAddr string `mapstructure:"listen_addr"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

type StorageConfig struct {
	R2    R2Config    `mapstructure:"r2"`
	Redis RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type R2Config struct {
	BucketName     string `mapstructure:"bucket_name"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	APIEndpoint    string `mapstructure:"api_endpoint"`
	PublicEndpoint string `mapstructure:"public_endpoint"`
}

type WorkerConfig struct {
	PoolSize int `mapstructure:"pool_size"`
}

type GCConfig struct {
	IntervalSeconds            int     `mapstructure:"interval_seconds"`
	LogRetentionDays           int     `mapstructure:"log_retention_days"`
	ContentSimilarityThreshold float64 `mapstructure:"content_similarity_threshold"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func LoadConfig(filename string) (*Config, error) {
	cfg := viper.New()
	cfg.SetConfigFile(filename)
	cfg.SetConfigType("yaml")

	cfg.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg.AutomaticEnv()

	cfg.SetDefault("app.name", "KnowLedger")
	cfg.SetDefault("app.dev", false)

	cfg.SetDefault("log.level", "info")
	cfg.SetDefault("log.output", []string{"stderr"})

	cfg.SetDefault("http.listen_addr", ":3000")

	cfg.SetDefault("worker.pool_size", 100)

	cfg.SetDefault("gc.interval_seconds", 60)
	cfg.SetDefault("gc.content_similarity_threshold", 0.85)
	cfg.SetDefault("gc.log_retention_days", 1)

	cfg.SetDefault("redis.addr", "localhost:6379")

	bindEnvs(cfg, Config{})

	if err := cfg.ReadInConfig(); err != nil {
		fmt.Printf("warning: config file not loaded (%v), using environment variables\n", err)
	}

	var config Config
	if err := cfg.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}

func bindEnvs(v *viper.Viper, iface interface{}, parts ...string) {
	t := reflect.TypeOf(iface)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag, ok := field.Tag.Lookup("mapstructure")
		if !ok || tag == "" {
			tag = strings.ToLower(field.Name)
		}

		path := append(parts, tag)

		if field.Type.Kind() == reflect.Struct {
			bindEnvs(v, reflect.New(field.Type).Elem().Interface(), path...)
			continue
		}

		_ = v.BindEnv(strings.Join(path, "."))
	}
}
