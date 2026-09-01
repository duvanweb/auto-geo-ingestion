package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"auto-geo-ingestion/internal/infrastructure/postgres/repositories"

	_ "github.com/lib/pq"
)

// NewConnection creates and configures a new PostgreSQL connection pool.
func NewConnection(config *Configuration) (repositories.Databaser, error) {
	dbURL := config.getDatabaseURL()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool settings.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func (c *Configuration) getDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName,
	)
}
