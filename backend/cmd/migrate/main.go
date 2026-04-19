package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"hrms/backend/internal/config"
	"hrms/backend/internal/db"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.OpenSQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	driver, err := newDriver(sqlDB)
	if err != nil {
		log.Fatal(err)
	}

	// For relative paths, "file://migrations" natively works in both Windows and Unix
	// without dealing with volume label syntax errors.
	sourceURL := "file://migrations"
	migrator, err := migrate.NewWithDatabaseInstance(sourceURL, cfg.DBName, driver)
	if err != nil {
		log.Fatal(err)
	}

	action := "up"
	if len(os.Args) > 1 {
		action = os.Args[1]
	}

	switch action {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Down()
	case "force":
		version := 0
		if len(os.Args) > 2 {
			if _, scanErr := fmt.Sscanf(os.Args[2], "%d", &version); scanErr != nil {
				log.Fatalf("invalid version for force: %v", scanErr)
			}
		}
		err = migrator.Force(version)
	default:
		log.Fatalf("unsupported migrate action %q", action)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	fmt.Printf("migration %s complete\n", action)
}

func newDriver(sqlDB *sql.DB) (database.Driver, error) {
	driver, err := postgresmigrate.WithInstance(sqlDB, &postgresmigrate.Config{
		SchemaName: "public",
	})
	if err != nil {
		return nil, err
	}
	return driver, nil
}
