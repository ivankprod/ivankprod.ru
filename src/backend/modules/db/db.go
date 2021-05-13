package db

import (
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Connect function
func Connect() (*sqlx.DB, error) {
	// Define database connection settings:
	maxConn, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
	maxIdleConn, _ := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNECTIONS"))
	maxLifetimeConn, _ := strconv.Atoi(os.Getenv("DB_MAX_LIFETIME_CONNECTIONS"))
	dbCredentials := fmt.Sprintf("%s:%s", os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))
	dbConnStr := fmt.Sprintf("tcp(%s:%s)/%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))

	// Define database connection:
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s@%s", dbCredentials, dbConnStr))
	if err != nil {
		return nil, err
	}

	// Set database connection settings:
	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(maxIdleConn)
	db.SetConnMaxLifetime(time.Duration(maxLifetimeConn))

	// Try to ping database:
	if err := db.Ping(); err != nil {
		defer db.Close()
		return nil, err
	}

	// Return normal connection
	return db, nil
}
