package metric

import (
	"context"
	"errors"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/filter"
	"github.com/joinself/restful-client/pkg/log"
)

// Repository encapsulates the logic to access metrics from the data source.
type Repository interface {
	// Count returns the number of metrics.
	Count(ctx context.Context, appID string, from, to int64) (int, error)
	// Query returns the list of metrics with the given offset and limit.
	Query(ctx context.Context, appid string, offset, limit int, from, to int64) ([]entity.Metric, error)
	// Upsert creates or updates an metric entry based on its uuid.
	Upsert(ctx context.Context, metric *entity.Metric) error
}

// repository persists metrics in database
type repository struct {
	db      *dbcontext.DB
	checker *filter.Checker
	logger  log.Logger
}

// NewRepository creates a new metric repository
func NewRepository(db *dbcontext.DB, checker *filter.Checker, logger log.Logger) Repository {
	return repository{db, checker, logger}
}

// Create saves a new metric record in the database.
// It returns the ID of the newly inserted metric record.
func (r repository) create(ctx context.Context, metric *entity.Metric) error {
	return r.db.With(ctx).Model(metric).Insert()
}

// Update saves the changes to an metric in the database.
func (r repository) update(ctx context.Context, metric entity.Metric) error {
	return r.db.With(ctx).Model(&metric).Update()
}

// Upsert creates or updates an metric entry based on its uuid.
func (r repository) Upsert(ctx context.Context, metric *entity.Metric) error {
	m, err := r.getByUUID(ctx, metric.AppID, metric.UUID)
	if err != nil {
		return r.create(ctx, metric)
	}
	metric.ID = m.ID
	return r.update(ctx, *metric)
}

// Count returns the number of the metric records in the database.
func (r repository) Count(ctx context.Context, appID string, from, to int64) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").
		From("metric").
		Where(&dbx.HashExp{"appid": appID}).
		AndWhere(dbx.Between("created_at", time.Unix(from, 0), time.Unix(to, 0))).
		Row(&count)
	return count, err
}

// Query retrieves the metric records with the specified offset and limit from the database.
func (r repository) Query(ctx context.Context, appid string, offset, limit int, from, to int64) ([]entity.Metric, error) {
	var metrics []entity.Metric
	err := r.db.With(ctx).
		Select().
		Where(&dbx.HashExp{"appid": appid}).
		AndWhere(dbx.Between("created_at", time.Unix(from, 0), time.Unix(to, 0))).
		OrderBy("id").
		Offset(int64(offset)).
		OrderBy("created_at DESC").
		Limit(int64(limit)).
		All(&metrics)
	return metrics, err
}

func (r repository) getByUUID(ctx context.Context, appID string, id int) (entity.Metric, error) {
	var metrics []entity.Metric

	err := r.db.With(ctx).
		Select().
		OrderBy("id").
		Where(&dbx.HashExp{"uuid": id, "appid": appID}).
		All(&metrics)

	if len(metrics) == 0 {
		return entity.Metric{}, errors.New("sql: no rows in result set")
	}

	return metrics[0], err
}
