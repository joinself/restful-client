package test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
	_ "github.com/mattn/go-sqlite3"
)

var db *dbcontext.DB

const CONFIG_RELATIVE_FILE = "../../config/test.yml"

// DB returns the database connection for testing purpose.
func DB(t *testing.T) *dbcontext.DB {
	if db != nil {
		return db
	}
	logger, _ := log.NewForTest()
	cf := fmt.Sprintf("%s/%s", getSourcePath(), CONFIG_RELATIVE_FILE)
	cfg, err := config.Load(cf, logger)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	err = os.MkdirAll(cfg.StorageDir, 0744)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	dbc, err := dbx.MustOpen("sqlite3", filepath.Join(cfg.StorageDir, "client.db"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	dbc.LogFunc = logger.Infof
	db = dbcontext.New(dbc)
	return db
}

// ResetTables truncates all data in the specified tables.
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	for _, table := range tables {
		q := `
			DELETE FROM ` + table + `;`
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
		ConnectionID: &connectionID,
	}).Insert()
}
