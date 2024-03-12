package metric

import (
	"net/http"
	"strconv"
	"time"

	"github.com/joinself/restful-client/pkg/log"
	"github.com/joinself/restful-client/pkg/pagination"
	"github.com/joinself/restful-client/pkg/response"
	"github.com/labstack/echo/v4"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(r *echo.Group, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.GET("/:app_id/metrics", res.query)
}

type resource struct {
	service Service
	logger  log.Logger
}

// ListMetrics godoc
// @Summary        Retrieve a paginated list of metrics
// @Description    Retrieves a paginated list of metrics for a specific app_id, matching the specified filters.
//
//	Pagination is provided through optional page and per_page parameters.
//	Defaults are page 1 and per_page 10 if not provided.
//
// @Tags           metrics
// @Accept         json
// @Produce        json
// @Security       BearerAuth
// @Param          app_id path string true "App's Unique Identifier (UUID)"
// @Param          page query int false "Page number for pagination, default is 1 if not provided."
// @Param          per_page query int false "Number of metrics per page for pagination, default is 10 if not provided."
// @Param          from query int false "Start time for the metrics query in Unix timestamp, default is 0 if not provided."
// @Param          to query int false "End time for the metrics query in Unix timestamp, default is current time if not provided."
// @Success        200 {object} ExtListResponse "Successful metrics list retrieval returns a 200 status and a list of metrics."
// @Failure        500 {object} response.Error "A 500 status and an error object will be returned in case of an internal server error during the request."
// @Router         /apps/{app_id}/metrics [get]
func (r resource) query(c echo.Context) error {
	ctx := c.Request().Context()
	var from, to int64
	from = 0
	to = time.Now().Unix()

	if v, err := strconv.Atoi(c.QueryParam("from")); err == nil {
		from = int64(v)
	}
	if v, err := strconv.Atoi(c.QueryParam("to")); err == nil {
		to = int64(v)
	}

	count, err := r.service.Count(ctx, c.Param("app_id"), from, to)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	pages := pagination.NewFromRequest(c.Request(), count)
	metrics, err := r.service.Query(ctx,
		c.Param("app_id"),
		pages.Offset(),
		pages.Limit(),
		from,
		to,
	)
	if err != nil {
		return c.JSON(response.DefaultInternalError(c, r.logger, err.Error()))
	}

	mms := []ExtMetric{}
	for _, m := range metrics {
		mms = append(mms, ExtMetric{
			ID:        m.UUID,
			Recipient: m.Recipient,
			Actions:   m.Actions,
			CreatedAt: m.CreatedAt,
		})
	}

	pages.Items = mms
	return c.JSON(http.StatusOK, pages)
}
