package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbhost := os.Getenv("DB_HOST")
	dbport := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, dbhost, dbport, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("db.ConnectDB: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("db.ConnectDB: %w", err)
	}
	return db, nil
}
