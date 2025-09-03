package data

import (
	"context"
	"embed"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
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

// DownAll rolls back all migrations.
func DownAll(dsn string) error {
	fsDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}
	migrateClient, err := migrate.NewWithSourceInstance("iofs", fsDriver, dsn)
	if err != nil {
		return err
	}
	if err = migrateClient.Down(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func NewSchema(dsn string) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	// Prompt user for new schema name
	var newSchemaName string
	fmt.Print("Enter the new schema name: ")
	_, err = fmt.Scan(&newSchemaName)
	if err != nil {
		panic(err)
	}
	_, err = pool.Exec(context.Background(), fmt.Sprintf("CREATE SCHEMA %s;", newSchemaName))
	if err != nil {
		panic(err)
	}
}

// Drops everything in the db
func DropEverything(dsn string) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()
	// Get all schemas from DSN
	var schemas []string
	rows, err := pool.Query(context.Background(), "SELECT schema_name FROM information_schema.schemata;")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			panic(err)
		}
		schemas = append(schemas, schema)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	// Ask user which schema to drop
	var schemaToDrop string
	fmt.Print("Enter the schema name to drop: ")
	_, err = fmt.Scan(&schemaToDrop)
	if err != nil {
		panic(err)
	}
	_, err = pool.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE;", schemaToDrop))
	if err != nil {
		panic(err)
	}
}

func Force(dsn string, version int) error {
	fsDriver, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return err
	}
	migrateClient, err := migrate.NewWithSourceInstance("iofs", fsDriver, dsn)
	if err != nil {
		return err
	}
	if err = migrateClient.Force(version); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func DbMigrationTool(dsn string) {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		panic("Failed to read embedded migrations: " + err.Error())
	}
	fmt.Println("Running database version check")
	versionNameMap := make(map[int]string)
	for _, entry := range entries {
		if !entry.IsDir() {
			version, err := strconv.Atoi(strings.Split(entry.Name(), "_")[0])
			if err == nil {
				versionNameMap[version] = entry.Name()
			}
		}
	}
	fmt.Println("Available database versions:")
	for version, name := range versionNameMap {
		fmt.Printf(" - %d: %s\n", version, name)
	}
	// Ask for user input to select a version
	var selectedVersion int
	fmt.Print("Enter the version number to migrate to: ")
	_, err = fmt.Scan(&selectedVersion)
	if err != nil {
		fmt.Println("Invalid input: " + err.Error())
	} else {
		err = Force(dsn, selectedVersion)
		if err != nil {
			fmt.Println("Failed to migrate database to version " + fmt.Sprintf("%d", selectedVersion) + ": " + err.Error())
		} else {
			fmt.Println("Successfully migrated database to version " + fmt.Sprintf("%d", selectedVersion))
		}
	}
}
