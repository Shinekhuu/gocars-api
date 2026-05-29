package main

import (
	"flag"
	"fmt"

	"gocars-api/internal/app"
	"gocars-api/internal/cache/redis"
	"gocars-api/internal/config"
	pgdb "gocars-api/internal/database/postgres"
	"gocars-api/internal/logger"
	"gocars-api/internal/search/meili"
	"gocars-api/internal/server"
	"gocars-api/scripts"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.MODE)

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: cfg.SENTRY_DSN,
	}); err != nil {
		zap.L().Warn("Sentry initialization failed", zap.Error(err))
	}

	// Only For mySQL (if needed in the future)
	// mysqldb.InitDB(cfg)
	pgdb.InitDB(cfg)
	redis.Init(cfg)
	meili.Init(cfg.MEILI_URL, cfg.MEILI_MASTER_KEY)

	runSync := flag.Bool("commands-sync", false, "Run sync command")
	flag.Parse()

	if *runSync {
		fmt.Println("Running Sync...")
		scripts.SyncLanguage()
		return
	}

	application := app.NewApp(pgdb.DB)
	server.Run(cfg, application)
}
