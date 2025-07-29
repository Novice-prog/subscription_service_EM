package storage

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"

	"log"
	"os"
	"strings"
	"time"
)

func InitDB() (*sqlx.DB, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_NAME", "subscriptions"),
	)

	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to database: %v", err)
	}

	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("Error running migrations: %w", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func runMigrations(db *sqlx.DB) error {
	files, err := os.ReadDir("./migrations")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		path := fmt.Sprintf("./migrations/%s", file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to exec %s: %w", file.Name(), err)
		}
		log.Printf("Applied migration: %s", file.Name())
	}
	return nil
}
