package models

import (
	"database/sql"

	"github.com/gobuffalo/packr/v2"
	"github.com/juju/loggo"
	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

var DB *sql.DB
var logger *loggo.Logger

func CloseDB() {
	DB.Close()

	return
}

func InitDB(connectionString string) {
	newLogger := loggo.GetLogger("web.models")
	logger = &newLogger

	logger.Debugf("Connecting to Database")
	dbClient, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Criticalf("Coud not connect to database: %s", err)
		panic(err)
	}
	DB = dbClient

	DB.SetMaxIdleConns(5)

	// Do Migration
	logger.Debugf("Loading Migrations")
	migrate.SetTable("web_migrations")
	migrations := &migrate.PackrMigrationSource{
		Box: packr.New("migrations","./migrations"),
	}

	logger.Debugf("Applying Migrations")
	n, err := migrate.Exec(DB, "postgres", migrations, migrate.Up)
	if n > 0 {
		logger.Infof("Applied %d migrations!\n", n)
	}
	if err != nil {
		logger.Criticalf("Coud not migrate database: %s", err)
		panic(err)
	}

	return
}
