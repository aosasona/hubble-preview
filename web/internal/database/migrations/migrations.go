package migrations

import (
	"embed"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)

//go:embed *.sql
var migrations embed.FS

var sourceDriver source.Driver

func init() {
	var err error
	sourceDriver, err = iofs.New(migrations, ".")
	if err != nil {
		panic(fmt.Sprintf("failed to create source driver: %v", err))
	}
}

type Direction int

const (
	Down Direction = iota
	Up
)

func Migrate(dsn string, direction Direction) error {
	dsn = strings.Replace(dsn, "postgres://", "pgx5://", 1)
	migrator, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	log.Info().Msg("running migrations")
	if direction == Down {
		return migrator.Down()
	}

	return migrator.Up()
}
