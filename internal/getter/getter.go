package getter

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"public-ip-getter/internal/config"
	"strings"
	"time"
)

type (
	Getter struct {
		cfg        *config.Config
		mux        *http.ServeMux
		httpServer *http.Server
	}
)

func New(cfg *config.Config) *Getter {
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:           cfg.ListenAddr,
		Handler:        mux,
		ReadTimeout:    time.Second * 2,
		WriteTimeout:   time.Second * 2,
		IdleTimeout:    time.Second * 30,
		MaxHeaderBytes: 512,
	}

	getter := &Getter{
		cfg:        cfg,
		mux:        mux,
		httpServer: httpServer,
	}

	getter.mux.Handle("/", getter)
	return getter
}

func (g *Getter) Run() error {
	slog.Info("start HTTP server", slog.String("listen-address", g.cfg.ListenAddr))
	err := g.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (g *Getter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ipStr string
	)

	// get ip from header
	if forwardHeader := r.Header.Get("X-Forwarded-For"); forwardHeader != "" {
		ip := net.ParseIP(strings.TrimSpace(strings.Split(forwardHeader, ",")[0]))
		if ip != nil {
			ipStr = ip.String()
		}
	}

	if ipStr == "" {
		if tmpIPStr, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
			slog.Error("parse remote addr error", slog.String("error", err.Error()))
		} else {
			ipStr = tmpIPStr
		}
	}

	if ipStr == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body := []byte(ipStr)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	n, err := w.Write(body)
	if err != nil {
		slog.Error("write HTTP response error", slog.String("error", err.Error()))
	} else {
		if n != len(body) {
			slog.Error(
				"write HTTP response failed, unexpected written size",
				slog.Int("expect", len(body)),
				slog.Int("actual", n),
			)
		}
	}
}

func (g *Getter) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()
	return g.httpServer.Shutdown(ctx)
}
