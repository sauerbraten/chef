package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/sauerbraten/chef"
	"github.com/sauerbraten/chef/db"
)

type Config struct {
	DatabaseFilePath    string        `env:"DB_FILE_PATH"       envDefault:"chef.sqlite"`
	ScanInterval        time.Duration `env:"SCAN_INTERVAL"      envDefault:"60s"`
	MasterServerAddress string        `env:"MASTER_SERVER_ADDR" envDefault:"master.sauerbraten.org:28787"`
	ListenAddress       string        `env:"LISTEN_ADDR"        envDefault:"localhost:8082"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	cfg := env.Must(env.ParseAs[Config]())

	// open SQLite DB
	db, err := db.New(cfg.DatabaseFilePath)
	if err != nil {
		slog.Error("open database", "error", err)
		os.Exit(1)
	}

	// start collector
	coll := chef.NewCollector(db, cfg.MasterServerAddress, cfg.ScanInterval)
	go coll.Run()

	// listen for SIGINT/SIGTERM for graceful web server shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// start web server
	slog.Info("serving web UI and JSON API", "listen_addr", cfg.ListenAddress)
	server := startWebserver(db, cfg.ListenAddress, stop)

	// wait for interrupt signal, then shut down webserver
	sig, ok := <-stop
	if !ok {
		slog.Info("exiting after web server error")
		db.Close()
		os.Exit(1)
	} else {
		slog.Info("received stop signal, gracefully shutting down", "signal", sig)
		err := server.Shutdown(context.Background())
		if err != nil {
			slog.Error("shutdown web server", "error", err)
		}
	}

	// clean up DB
	err = db.Close()
	if err != nil {
		slog.Error("close DB connection", "error", err)
	}

	slog.Info("good bye!")
}

func startWebserver(db *db.Database, listenAddr string, stop chan os.Signal) *http.Server {
	r := http.NewServeMux()
	r.Handle("/", chef.NewWebUI(db))
	r.Handle("/api/", http.StripPrefix("/api", chef.NewAPI(db)))

	s := &http.Server{
		Addr:    listenAddr,
		Handler: withRequestLogging(r),
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("serve web UI", "error", err)
			close(stop) // unblock main() goroutine
		}
	}()

	return s
}

func withRequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ips := strings.Join(req.Header.Values("X-Forwarded-For"), ",")
		clientIP := strings.Split(ips, ",")[0]
		clientIP = strings.TrimSpace(clientIP)
		if clientIP == "" {
			clientIP = req.RemoteAddr
		}

		slog.Info("handling request",
			"request_method", req.Method, "request_url", req.URL.String(), "remote_addr", clientIP)

		next.ServeHTTP(resp, req)
	})
}
