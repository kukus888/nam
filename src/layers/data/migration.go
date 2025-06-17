package data

import (
	"embed"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*
var migrationFS embed.FS

func AutoMigrate(dsn string) error {
	fsDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}
	migrateClient, err := migrate.NewWithSourceInstance("iofs", fsDriver, dsn)
	if err != nil {
		return err
	}
	if err = migrateClient.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
