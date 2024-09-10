package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server        Server   `mapstructure:"server"`
	Postgres      Postgres `mapstructure:"postgres"`
	LogLevel      int      `mapstructure:"loglevel"`
	ServerTimeout int      `mapstructure:"server_timeout"`
	//TokenID       string   `mapstructure:"token_id"`
	Secret string `mapstructure:"secret"`
}

type Postgres struct {
	ConnectTimeout int `mapstructure:"connect_timeout"`
	Server         `mapstructure:"pg_server"`
	DB             string `mapstructure:"db"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
}

type Server struct {
	Path string `mapstructure:"path"`
	Port int    `mapstructure:"port"`
}

func (s *Server) ToString() string {
	return fmt.Sprintf("%s:%d", s.Path, s.Port)
}

func Load() (*Config, error) {
	path, name := fetchConfigPaht()
	if path == "" || name == "" {
		return nil, fmt.Errorf("config load error: %w", errors.New("path or name is empty"))
	}

	var cfg Config

	cfg, err := initViper(path, name, cfg)
	if err != nil {
		return nil, fmt.Errorf("config load error: %w", err)
	}

	var level slog.Level = slog.Level(cfg.LogLevel)
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	slog.SetDefault(log)

	return &cfg, nil
}

func initViper(path, name string, cfg Config) (Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("init viper ReadInConfig() err: %w", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("init viper Unmarshal() err: %w", err)
	}

	return cfg, nil
}

func fetchConfigPaht() (path string, name string) {

	flag.StringVar(&path, "config_path", "", "path to config file")
	flag.StringVar(&name, "config_name", "", "config file name")

	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	if name == "" {
		name = os.Getenv("CONFIG_NAME")
	}

	return path, name
}
