package test

import (
	"context"
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
	_ "github.com/lib/pq" // initialize posgresql for test
)

var db *dbcontext.DB

// DB returns the database connection for testing purpose.
func DB(t *testing.T) *dbcontext.DB {
	if db != nil {
		return db
	}
	logger, _ := log.NewForTest()
	cf := fmt.Sprintf("%s/../../config/%s.yml", getSourcePath(), getEnv("ENV", "local"))
	cfg, err := config.Load(cf, logger)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dbc, err := dbx.MustOpen("postgres", cfg.DSN)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dbc.LogFunc = logger.Infof
	db = dbcontext.New(dbc)
	return db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// ResetTables truncates all data in the specified tables.
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	for _, table := range tables {
		q := `
			DELETE FROM ` + table + `;
			TRUNCATE TABLE ` + table + ` CASCADE;`
		_, err := db.DB().NewQuery(q).Execute()
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
	}
}

// getSourcePath returns the directory containing the source code that is calling this function.
func getSourcePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

// CreateConnection creates a connection.
func CreateConnection(ctx context.Context, db *dbcontext.DB, id int) error {
	var connection entity.Connection
	err := db.With(ctx).Select().Model(id, &connection)
	if err == nil {
		return nil
	}

	return db.With(ctx).Model(&entity.Connection{
		ID:     id,
		AppID:  "app_" + strconv.Itoa(id),
		SelfID: "connection_" + strconv.Itoa(id),
		Name:   "connection_" + strconv.Itoa(id),
	}).Insert()
}

func CreateRequest(ctx context.Context, db *dbcontext.DB, id string, connectionID int) error {
	var request entity.Request
	err := db.With(ctx).Select().Model(id, &request)
	if err == nil {
		return nil
	}

	return db.With(ctx).Model(&entity.Request{
		ID:           id,
		ConnectionID: connectionID,
	}).Insert()
}
