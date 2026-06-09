package postgres

import (
	"context"
	"fmt"
	"net"
	"time"

	"gocars-api/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg config.Config) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.PG_HOST,
		cfg.PG_USER,
		cfg.PG_PASSWORD,
		cfg.PG_NAME,
		cfg.PG_PORT,
		cfg.PG_SSL_MODE,
	)

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		zap.L().Fatal("failed to parse PostgreSQL config", zap.Error(err))
	}

	// Force IPv4 — Docker containers may resolve to IPv6 which is unreachable
	pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp4", addr)
	}

	// Disable prepared statements — required for Supabase/PgBouncer in transaction mode.
	// PgBouncer routes each transaction to a different backend, so session-scoped
	// prepared statements collide across requests (SQLSTATE 42P05).
	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	sqlDB := stdlib.OpenDB(*pgxConfig)

	var gormDB *gorm.DB
	gormDB, err = gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		zap.L().Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}

	DB = gormDB

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	zap.L().Info("PostgreSQL initialized successfully")
}
