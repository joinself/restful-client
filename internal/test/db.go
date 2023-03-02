package test

import (
	"context"
	"path"
	"runtime"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/google/uuid"
	_ "github.com/lib/pq" // initialize posgresql for test
	"github.com/qiangxue/go-rest-api/internal/config"
	"github.com/qiangxue/go-rest-api/internal/entity"
	"github.com/qiangxue/go-rest-api/pkg/dbcontext"
	"github.com/qiangxue/go-rest-api/pkg/log"
)

var db *dbcontext.DB

// DB returns the database connection for testing purpose.
func DB(t *testing.T) *dbcontext.DB {
	if db != nil {
		return db
	}
	logger, _ := log.NewForTest()
	dir := getSourcePath()
	cfg, err := config.Load(dir+"/../../config/local.yml", logger)
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

// ResetTables truncates all data in the specified tables.
func ResetTables(t *testing.T, db *dbcontext.DB, tables ...string) {
	for _, table := range tables {
		q := `SET FOREIGN_KEY_CHECKS = 0;
		TRUNCATE TABLE '` + table + `';
		SET FOREIGN_KEY_CHECKS = 0;`
		err := db.DB().NewQuery(q).LastError
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		_, _ = db.DB().TruncateTable(table).Execute()
	}
}

// getSourcePath returns the directory containing the source code that is calling this function.
func getSourcePath() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func CreateConnection(ctx context.Context, db *dbcontext.DB, id string) error {
	var connection entity.Connection
	err := db.With(ctx).Select().Model(id, &connection)
	if err == nil {
		return nil
	}

	return db.With(ctx).Model(&entity.Connection{
		ID:     id,
		Name:   "connection_" + id,
		Selfid: uuid.New().String(),
	}).Insert()
}
