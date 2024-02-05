package clean

import (
	"context"
	"fmt"

	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
)

type Service interface {
	Clean()
}

type Config struct {
	DB     *dbcontext.DB
	Period int
	Tables []string
	Logger log.Logger
}

type service struct {
	db     *dbcontext.DB
	period int
	tables []string
	logger log.Logger
}

func NewService(c Config) Service {
	return &service{c.DB, c.Period, c.Tables, c.Logger}
}

func (s *service) Clean() {
	for _, t := range s.tables {
		err := s.cleanTable(t, s.period)
		if err != nil {
			s.logger.With(context.Background()).Info()
		}
	}
}

func (s *service) cleanTable(table string, period int) error {
	sql := `DELETE FROM %s WHERE created_at < datetime('now', '-%d days');`
	query := fmt.Sprintf(sql, table, period)
	_, err := s.db.DB().NewQuery(query).Execute()

	return err
}
