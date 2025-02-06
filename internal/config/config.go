package config

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
)

type (
	Config struct {
		ListenAddr string `json:"listenAddr"`
	}
)

func (c *Config) Validate() (err error) {
	defer func() {
		if err != nil {
			slog.Error(
				"bad config",
				slog.String("error", err.Error()),
			)
		}
	}()

	if c.ListenAddr == "" {
		err = errors.New("empty listen address")
		return
	}

	return
}

func Get() (*Config, error) {
	var (
		configFilePath *string
		file           *os.File
		bytes          []byte
		cfg            = &Config{}
	)

	configFilePath = flag.String("c", "./config.json", "config JSON file")
	flag.Parse()
	slog.Info("config file path", slog.String("file-path", *configFilePath))

	if f, err := os.Open(*configFilePath); err != nil {
		slog.Error("open config file error", slog.String("error", err.Error()))
		return nil, err
	} else {
		file = f
		defer func() {
			if file != nil {
				if e := file.Close(); e != nil {
					slog.Error("close config file error", slog.String("error", e.Error()))
				}
			}
		}()
	}

	if b, err := io.ReadAll(file); err != nil {
		slog.Error("read config file error", slog.String("error", err.Error()))
		return nil, err
	} else {
		bytes = b
	}

	if err := json.Unmarshal(bytes, cfg); err != nil {
		slog.Error("unmarshal config file error", slog.String("error", err.Error()))
		return nil, err
	}

	return cfg, nil
}
