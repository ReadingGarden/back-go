package db_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ReadingGarden/back-go/internal/config"
	"github.com/ReadingGarden/back-go/internal/db"
)

func TestBuildDSN(t *testing.T) {
	t.Parallel()

	cfg := config.DatabaseConfig{
		User:            "reader",
		Password:        "secret",
		Host:            "localhost",
		Port:            "3306",
		Name:            "reading_garden",
		MaxOpenConns:    15,
		MaxIdleConns:    4,
		ConnMaxLifetime: 2 * time.Minute,
	}

	dsn := db.BuildDSN(cfg)

	wants := []string{
		"reader:secret@tcp(localhost:3306)/reading_garden",
		"parseTime=true",
		"loc=Local",
		"charset=utf8mb4",
		"multiStatements=false",
	}
	for _, want := range wants {
		if !strings.Contains(dsn, want) {
			t.Fatalf("dsn %q does not contain %q", dsn, want)
		}
	}
}

func TestApplyPool(t *testing.T) {
	t.Parallel()

	cfg := config.DatabaseConfig{
		MaxOpenConns:    12,
		MaxIdleConns:    3,
		ConnMaxLifetime: 90 * time.Second,
	}

	sqlDB := sql.OpenDB(testConnector{})
	db.ApplyPool(sqlDB, cfg)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 12 {
		t.Fatalf("MaxOpenConnections = %d, want %d", stats.MaxOpenConnections, 12)
	}
}

type testConnector struct{}

func (testConnector) Connect(_ context.Context) (driver.Conn, error) {
	return testConn{}, nil
}

func (testConnector) Driver() driver.Driver {
	return testDriver{}
}

type testDriver struct{}

func (testDriver) Open(_ string) (driver.Conn, error) {
	return testConn{}, nil
}

type testConn struct{}

func (testConn) Prepare(_ string) (driver.Stmt, error) {
	return testStmt{}, nil
}

func (testConn) Close() error {
	return nil
}

func (testConn) Begin() (driver.Tx, error) {
	return testTx{}, nil
}

type testStmt struct{}

func (testStmt) Close() error {
	return nil
}

func (testStmt) NumInput() int {
	return 0
}

func (testStmt) Exec(_ []driver.Value) (driver.Result, error) {
	return testResult{}, nil
}

func (testStmt) Query(_ []driver.Value) (driver.Rows, error) {
	return testRows{}, nil
}

type testTx struct{}

func (testTx) Commit() error {
	return nil
}

func (testTx) Rollback() error {
	return nil
}

type testResult struct{}

func (testResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (testResult) RowsAffected() (int64, error) {
	return 0, nil
}

type testRows struct{}

func (testRows) Columns() []string {
	return []string{"value"}
}

func (testRows) Close() error {
	return nil
}

func (testRows) Next(_ []driver.Value) error {
	return io.EOF
}
