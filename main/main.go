package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"public-ip-getter/internal/config"
	"public-ip-getter/internal/getter"
	"syscall"
)

func main() {
	var (
		cfg    *config.Config
		stopCh = make(chan os.Signal, 1)
		s      os.Signal
	)

	// config
	if c, err := config.Get(); err != nil {
		panic(err)
	} else {
		if e := c.Validate(); e != nil {
			panic(e)
		}
		cfg = c
		slog.Info("config", slog.String("value", fmt.Sprintf("%+v", *cfg)))
	}

	// run
	g := getter.New(cfg)
	defer func() {
		if err := g.Shutdown(); err != nil {
			slog.Error("shutdown HTTP server error", slog.String("error", err.Error()))
		} else {
			slog.Info("shutdown HTTP server success")
		}
	}()
	go func() {
		if err := g.Run(); err != nil {
			slog.Error("run HTTP server error", slog.String("error", err.Error()))
		}
	}()

	// block
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(stopCh)
		close(stopCh)
	}()
	s = <-stopCh
	slog.Info("stop signal received", slog.String("signal", s.String()))
}
