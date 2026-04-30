package db

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ConnectDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func RunMigrations(db *gorm.DB, migrationsPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := mysqlmigrate.WithInstance(sqlDB, &mysqlmigrate.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "mysql", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
