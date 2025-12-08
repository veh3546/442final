package data_access

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB is a package-level *sql.DB that other packages can reference if needed.
var DB *sql.DB

// NewDB opens a MySQL connection pool and verifies connectivity.
func NewDB(user, pass, host, port, name string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		user, pass, host, port, name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Pool settings - tune these for your environment
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Verify connectivity with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	DB = db
	return db, nil
}
