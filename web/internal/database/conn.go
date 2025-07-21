package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var PgTypes = []string{
	"plugin_mode",
}

func InitializePool(dsn string, debug bool) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		for _, tName := range PgTypes {
			t, err := conn.LoadType(ctx, tName)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(t)

			t, err = conn.LoadType(ctx, "_"+tName)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(t)
		}

		return nil
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
