package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

const SSLModeDisable SSLMode = "disable"

type SSLMode string

func ConnectAndMigrate(
	host,
	port,
	databaseName,
	user,
	password string,
	sslMode SSLMode,
) error {

	connectionStr := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host,
		port,
		databaseName,
		user,
		password,
		sslMode,
	)

	db, err := sqlx.Open("pgx", connectionStr)

	if err != nil {
		return fmt.Errorf(
			"failed to open database connection: %w",
			err,
		)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()

	if err != nil {
		return fmt.Errorf(
			"database ping failed: %w",
			err,
		)
	}

	log.Println("Database connected successfully")

	DB = db

	return migrateUp(db)
}

func migrateUp(db *sqlx.DB) error {

	log.Println("Starting database migrations...")

	driver, err := postgres.WithInstance(
		db.DB,
		&postgres.Config{},
	)

	if err != nil {
		return fmt.Errorf(
			"failed to create migration driver: %w",
			err,
		)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres",
		driver,
	)

	if err != nil {
		return err
	}

	err = m.Up()

	if err != nil {

		if errors.Is(err, migrate.ErrNoChange) {

			log.Println("No new migrations to apply")

			return nil
		}

		return fmt.Errorf(
			"migration failed: %w",
			err,
		)
	}

	log.Println("Migrations applied successfully")

	return nil
}

func WithTransaction(
	ctx context.Context,
	fn func(*sqlx.Tx) error,
) error {

	tx, err := DB.BeginTxx(ctx, nil)

	if err != nil {
		return fmt.Errorf(
			"failed to begin transaction: %w",
			err,
		)
	}

	defer func() {

		if p := recover(); p != nil {

			_ = tx.Rollback()

			panic(p)
		}
	}()

	err = fn(tx)

	if err != nil {

		_ = tx.Rollback()

		return err
	}

	return tx.Commit()
}
