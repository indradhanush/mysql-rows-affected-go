package main

import (
	"database/sql"
	"fmt"
	"path"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pkg/errors"
)

func TestUserStore_Upsert(t *testing.T) {
	// Sets up a test database and applies the migrations. Out of scope for this blog post.
	db, err := setupDB(t)
	if err != nil {
		t.Fatal("failed to setup test database", err)
	}

	store := UserStore{db: db}
	if err = store.Upsert("johnwick"); err != nil {
		t.Fatal("failed to insert", err)
	}

	err = store.Upsert("johnwick")
	if err != nil {
		t.Fatal("failed to update", err)
	}
}

func setupDB(t *testing.T) (*sql.DB, error) {
	// Assuming you already have a mysql server running at localhost:3306 with root user and no
	// password abcd1234.
	url := "root:abcd1234@tcp(localhost:3306)/"
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to mysql on %q", url)
	}
	defer db.Close()

	dbName := fmt.Sprintf("rows_affected_test%d_%s", time.Now().Unix(), t.Name())
	dbName = fmt.Sprintf("%.64s", dbName)

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s`", dbName))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create database %q", dbName)
	}

	dsn := path.Join(url, fmt.Sprintf("%s?parseTime=true&loc=UTC&multiStatements=true", dbName))
	dbConn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database %q", dbName)
	}

	// Ensure that the database is removed when the test exits.
	t.Cleanup(func() {
		dbConn.Exec(fmt.Sprintf("DROP DATABASE `%s`", dbName))
		dbConn.Close()
	})

	driver, err := mysqlmigrate.WithInstance(dbConn, &mysqlmigrate.Config{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to configure driver for database %q", dbName)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup migrations for database %q", dbName)
	}

	if err = m.Up(); err != nil {
		return nil, errors.Wrapf(err, "failed to apply migrations to database %q", dbName)
	}

	return dbConn, nil
}
