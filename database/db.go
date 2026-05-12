package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var Todo *sqlx.DB

const SSLModeDisable SSLMode = "disable"

type SSLMode string

func ConnectandMigrate(host, port, databaseName, user, password string, sslMode SSLMode) error {
	connectionStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, databaseName, user, password, sslMode)

	DB, err := sqlx.Open("pgx", connectionStr)
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("database ping failed %w", err)
	}

	fmt.Println("Database connected successfully")
	Todo = DB
	return migrateUp(DB)
}

func migrateUp(db *sqlx.DB) error {
	fmt.Println("Starting database migrations...")
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not connect to database: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://database/migrations",
		"postgres", driver)

	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}

func WithTransaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := Todo.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			err := tx.Rollback()
			if err != nil {
				return
			}
		}
	}()

	err = fn(tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

//DB - dbHelper (queries) - handler(funs) - server/routes
// models <- body
