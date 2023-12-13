package dbcontext

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	dbx "github.com/go-ozzo/ozzo-dbx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

const DSN = "sqlite3:///Users/adriancidrejugo/self/self-storage/client.db"

func TestNew(t *testing.T) {
	runDBTest(t, func(db *dbx.DB) {
		dbc := New(db)
		assert.NotNil(t, dbc)
		assert.Equal(t, db, dbc.DB())
	})
}

func TestDB_Transactional(t *testing.T) {
	runDBTest(t, func(db *dbx.DB) {
		assert.Zero(t, runCountQuery(t, db))
		dbc := New(db)

		// successful transaction
		err := dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "1", "name": "name1"}).Execute()
			assert.Nil(t, err)
			_, err = dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "2", "name": "name2"}).Execute()
			assert.Nil(t, err)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 2, runCountQuery(t, db))

		// failed transaction
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "3", "name": "name1"}).Execute()
			assert.Nil(t, err)
			_, err = dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "4", "name": "name2"}).Execute()
			assert.Nil(t, err)
			return sql.ErrNoRows
		})
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, 2, runCountQuery(t, db))

		// failed transaction, but queries made outside of the transaction
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			_, err := dbc.With(context.Background()).Insert("dbcontexttest", dbx.Params{"id": "3", "name": "name1"}).Execute()
			assert.Nil(t, err)
			_, err = dbc.With(context.Background()).Insert("dbcontexttest", dbx.Params{"id": "4", "name": "name2"}).Execute()
			assert.Nil(t, err)
			return sql.ErrNoRows
		})
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, 4, runCountQuery(t, db))
	})
}

/*
func TestDB_TransactionHandler(t *testing.T) {
	runDBTest(t, func(db *dbx.DB) {
		assert.Zero(t, runCountQuery(t, db))
		dbc := New(db)
		txHandler := dbc.TransactionHandler()

		// successful transaction
		{
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
			err := routing.NewContext(res, req, txHandler, func(c *routing.Context) error {
				ctx := c.Request.Context()
				_, err := dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "1", "name": "name1"}).Execute()
				assert.Nil(t, err)
				_, err = dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "2", "name": "name2"}).Execute()
				assert.Nil(t, err)
				return nil
			}).Next()
			assert.Nil(t, err)
			assert.Equal(t, 2, runCountQuery(t, db))
		}

		// failed transaction
		{
			res := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "http://127.0.0.1/users", nil)
			err := routing.NewContext(res, req, txHandler, func(c *routing.Context) error {
				ctx := c.Request.Context()
				_, err := dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "3", "name": "name1"}).Execute()
				assert.Nil(t, err)
				_, err = dbc.With(ctx).Insert("dbcontexttest", dbx.Params{"id": "4", "name": "name2"}).Execute()
				assert.Nil(t, err)
				return sql.ErrNoRows
			}).Next()
			assert.Equal(t, err, sql.ErrNoRows)
			assert.Equal(t, 2, runCountQuery(t, db))
		}
	})
}
*/

func runDBTest(t *testing.T, f func(db *dbx.DB)) {
	// TODO: Get this stuff from an env var...
	storageDir := "/Users/adriancidrejugo/self/self-storage/"
	db, err := dbx.MustOpen("sqlite3", filepath.Join(storageDir, "client.db"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer func() {
		_ = db.Close()
	}()

	sqls := []string{
		"CREATE TABLE IF NOT EXISTS dbcontexttest (id VARCHAR PRIMARY KEY, name VARCHAR)",
		"DELETE FROM dbcontexttest",
	}
	for _, s := range sqls {
		_, err = db.NewQuery(s).Execute()
		if err != nil {
			t.Error(err, " with SQL: ", s)
			t.FailNow()
		}
	}

	f(db)
}

func runCountQuery(t *testing.T, db *dbx.DB) int {
	var count int
	err := db.NewQuery("SELECT COUNT(*) FROM dbcontexttest").Row(&count)
	assert.Nil(t, err)
	return count

}
