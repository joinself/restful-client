package database

import (
	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/pkg/log"
)

func setupPostgres(dsn string, logger log.Logger) (*dbx.DB, error) {
	return dbx.MustOpen("postgres", dsn)
}
