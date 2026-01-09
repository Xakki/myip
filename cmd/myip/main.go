package main

import (
	"context"
	"io"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"

	"github.com/Graylog2/go-gelf/gelf"

	"myip/internal/config"
	"myip/internal/rdap"
	"myip/internal/store"
	"myip/internal/web"
)

const shutdownTimeout = 5 * time.Second

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := initLogger(cfg)

	redisStore := store.NewRedisStore(cfg.Redis, cfg.RedisUser, cfg.RedisPass)
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := redisStore.Ping(ctx); err != nil {
		logger.Fatalf("redis error: %v", err)
	}

	templates, err := web.ParseTemplates()
	if err != nil {
		logger.Fatalf("template error: %v", err)
	}

	var rdapClient web.RDAPLookup
	if cfg.RDAPAPI != "" {
		rdapClient = rdap.NewClient(cfg.RDAPAPI)
	}

	onError := func(err error) {
		if err == nil {
			return
		}
		logger.Printf("error: %v", err)
	}

	service := web.NewService(redisStore, rdapClient, onError)
	handler := web.NewHandler(templates, service)

	server := &http.Server{
		Addr:              cfg.WebAddr,
		Handler:           handler,
		ReadHeaderTimeout: 3 * time.Second,
	}

	logger.Printf("listening on %s", cfg.WebAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("server error: %v", err)
	}
}

func initLogger(cfg config.Config) *log.Logger {
	var writer io.Writer
	var err error

	switch cfg.LogType {
	case "syslog":
		writer, err = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "myip")
		if err != nil {
			log.Printf("failed to initialize syslog: %v, falling back to stderr", err)
			writer = os.Stderr
		}
	case "gelf":
		if cfg.LogAddr == "" {
			log.Printf("GELF address is not specified, falling back to stderr")
			writer = os.Stderr
		} else {
			writer, err = gelf.NewWriter(cfg.LogAddr)
			if err != nil {
				log.Printf("failed to initialize GELF: %v, falling back to stderr", err)
				writer = os.Stderr
			}
		}
	case "system":
		// On Linux, this might be journald, but through syslog or stderr.
		// Here we'll treat "system" as syslog with fallback to stderr.
		writer, err = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "myip")
		if err != nil {
			writer = os.Stderr
		}
	case "console":
		fallthrough
	default:
		writer = os.Stderr
	}

	flags := log.LstdFlags
	if cfg.LogType == "syslog" || cfg.LogType == "system" || cfg.LogType == "gelf" {
		flags = 0 // Remote loggers usually handle timestamps themselves
	}

	return log.New(writer, "", flags)
}
