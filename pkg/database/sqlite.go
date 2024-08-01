package database

import (
	"os"
	"path/filepath"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/pkg/log"
)

func setupSQLite(dsn string, logger log.Logger) (*dbx.DB, error) {
	err := os.MkdirAll(filepath.Dir(dsn), 0744)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	db, err := dbx.MustOpen("sqlite3", dsn)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	return db, nil
}
