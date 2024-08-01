package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/pkg/log"
)

func SetupDB(cfg *config.Config, logger log.Logger) (*dbx.DB, error) {
	parsedURL, err := url.Parse(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("invalid DSN: %v", err)
	}

	var db *dbx.DB
	switch parsedURL.Scheme {
	case "postgres":
		db, err = setupPostgres(cfg.DSN, logger)
	case "sqlite3":
		dsnWithoutScheme := strings.TrimPrefix(cfg.DSN, "sqlite3://")
		db, err = setupSQLite(dsnWithoutScheme, logger)
	default:
		return nil, fmt.Errorf("unsupported DSN scheme: %s", parsedURL.Scheme)
	}

	db.QueryLogFunc = logDBQuery(logger)
	db.ExecLogFunc = logDBExec(logger)

	return db, err
}

// logDBQuery returns a logging function that can be used to log SQL queries.
func logDBQuery(logger log.Logger) dbx.QueryLogFunc {
	return func(ctx context.Context, t time.Duration, sql string, rows *sql.Rows, err error) {
		if err == nil {
			logger.With(ctx, "duration", t.Milliseconds(), "sql", sql).Info("DB query successful")
		} else {
			logger.With(ctx, "sql", sql).Errorf("DB query error: %v", err)
		}
	}
}

// logDBExec returns a logging function that can be used to log SQL executions.
func logDBExec(logger log.Logger) dbx.ExecLogFunc {
	return func(ctx context.Context, t time.Duration, sql string, result sql.Result, err error) {
		if err == nil {
			logger.With(ctx, "duration", t.Milliseconds(), "sql", sql).Info("DB execution successful")
		} else {
			logger.With(ctx, "sql", sql).Errorf("DB execution error: %v", err)
		}
	}
}
