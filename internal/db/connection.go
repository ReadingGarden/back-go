package db

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/ReadingGarden/back-go/internal/config"
)

func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", BuildDSN(cfg))
	if err != nil {
		return nil, err
	}

	ApplyPool(db, cfg)

	return db, nil
}

func BuildDSN(cfg config.DatabaseConfig) string {
	dsn := mysql.NewConfig()
	dsn.User = cfg.User
	dsn.Passwd = cfg.Password
	dsn.Net = "tcp"
	dsn.Addr = cfg.Host + ":" + cfg.Port
	dsn.DBName = cfg.Name
	dsn.ParseTime = true
	dsn.Loc = time.Local
	dsn.Params = map[string]string{
		"charset":         "utf8mb4",
		"collation":       "utf8mb4_unicode_ci",
		"multiStatements": "false",
	}

	return dsn.FormatDSN()
}

func ApplyPool(db *sql.DB, cfg config.DatabaseConfig) {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
}
