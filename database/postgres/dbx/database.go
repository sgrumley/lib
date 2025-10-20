package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDBConnForApp(log *slog.Logger, cc *ConnectionConfig, dsc *SQLConfig) (*sql.DB, error) {
	dsn := URLForConfig(*cc)
	log.Info("DB", "Connection", cc)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	if dsc != nil {
		if dsc.MaxOpenConns > 0 {
			cfg.MaxConns = int32(dsc.MaxOpenConns)
		}
		if dsc.MaxIdleConns > 0 {
			cfg.MinConns = int32(dsc.MaxIdleConns)
		}
		if dsc.ConnMaxLifetime > 0 {
			cfg.MaxConnLifetime = dsc.ConnMaxLifetime
		}

		log.Info("DB connection data",
			slog.Group("connection pools",
				slog.Int("max open connections", dsc.MaxOpenConns),
				slog.Int("max idle connections", dsc.MaxIdleConns),
				slog.Duration("connection max life time", dsc.ConnMaxLifetime),
			),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool connect failed: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	return pool, nil
}
