package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	HTTP          HTTPConfig          `mapstructure:"http"`
	GRPC          GRPCConfig          `mapstructure:"grpc"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Postgres      PostgresConfig      `mapstructure:"postgres"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Secret  string `mapstructure:"secret"`
	Version string `mapstructure:"version"`
}

type HTTPConfig struct {
	Port            string        `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type GRPCConfig struct {
	Port            string        `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            string        `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

type ObservabilityConfig struct {
	LogLevel string `mapstructure:"log_level"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	err := v.ReadInConfig()
	if err != nil {
		// Allow missing file when env vars / defaults are enough.
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	bindEnv(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "github.com/VH288/tc-template")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.version", "0.1.0")
	v.SetDefault("http.port", "8081")
	v.SetDefault("http.read_timeout", "15s")
	v.SetDefault("http.write_timeout", "15s")
	v.SetDefault("http.shutdown_timeout", "10s")
	v.SetDefault("grpc.port", "7000")
	v.SetDefault("grpc.shutdown_timeout", "10s")
	v.SetDefault("postgres.host", "127.0.0.1")
	v.SetDefault("postgres.port", "5432")
	v.SetDefault("postgres.user", "root")
	v.SetDefault("postgres.password", "password123")
	v.SetDefault("postgres.dbname", "sample")
	v.SetDefault("postgres.sslmode", "disable")
	v.SetDefault("postgres.max_open_conns", 25)
	v.SetDefault("postgres.max_idle_conns", 5)
	v.SetDefault("postgres.conn_max_lifetime", "30m")
	v.SetDefault("observability.log_level", "info")
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("app.name", "APP_NAME")
	_ = v.BindEnv("app.secret", "APP_SECRET")
	_ = v.BindEnv("http.port", "PORT", "HTTP_PORT")
	_ = v.BindEnv("grpc.port", "GRPC_PORT")
	_ = v.BindEnv("postgres.host", "DB_HOST")
	_ = v.BindEnv("postgres.port", "DB_PORT")
	_ = v.BindEnv("postgres.user", "DB_USER")
	_ = v.BindEnv("postgres.password", "DB_PASSWORD")
	_ = v.BindEnv("postgres.dbname", "DB_NAME")
}
