package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	lg "log"
	"net/http"
	"os"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/content"
	"github.com/go-ozzo/ozzo-routing/v2/cors"
	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/errors"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/healthcheck"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/internal/self"
	"github.com/joinself/restful-client/pkg/accesslog"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
	_ "github.com/lib/pq"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

func main() {
	flag.Parse()
	// create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// connect to the database
	db, err := dbx.MustOpen("postgres", cfg.DSN)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	db.QueryLogFunc = logDBQuery(logger)
	db.ExecLogFunc = logDBExec(logger)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// build HTTP server
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	hs := &http.Server{
		Addr:    address,
		Handler: buildHandler(logger, dbcontext.New(db), cfg),
	}

	// start the HTTP server with graceful shutdown
	go routing.GracefulShutdown(hs, 10*time.Second, logger.Infof)
	logger.Infof("server %v is running at %v", Version, address)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		os.Exit(-1)
	}
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB, cfg *config.Config) http.Handler {
	client, err := setupSelfClient(cfg)
	if err != nil {
		lg.Fatalf("failed to setup self client: %v", err.Error())
	}

	router := routing.New()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		content.TypeNegotiator(content.JSON),
		cors.Handler(cors.AllowAll),
	)

	healthcheck.RegisterHandlers(router, Version)

	rg := router.Group("/v1")

	authHandler := auth.Handler(cfg.JWTSigningKey)
	connectionRepo := connection.NewRepository(db, logger)
	messageRepo := message.NewRepository(db, logger)
	factRepo := fact.NewRepository(db, logger)
	attestationRepo := attestation.NewRepository(db, logger)

	connection.RegisterHandlers(rg.Group(""),
		connection.NewService(connectionRepo, logger, client.FactService()),
		authHandler, logger,
	)

	message.RegisterHandlers(rg.Group(""),
		message.NewService(messageRepo, logger, client),
		authHandler, logger,
	)

	fact.RegisterHandlers(rg.Group(""),
		fact.NewService(factRepo, attestationRepo, logger, client.FactService()),
		authHandler, logger,
	)

	auth.RegisterHandlers(rg.Group(""),
		auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, logger),
		logger,
	)

	self.RegisterHandlers(
		self.NewService(client, connectionRepo, factRepo, messageRepo, logger),
		logger,
	)

	return router
}

func setupSelfClient(cfg *config.Config) (*selfsdk.Client, error) {
	return selfsdk.New(selfsdk.Config{
		SelfAppID:           cfg.SelfAppID,
		SelfAppDeviceSecret: cfg.SelfAppDeviceSecret,
		StorageKey:          cfg.SelfStorageKey,
		StorageDir:          cfg.SelfStorageDir,
		Environment:         cfg.SelfEnv,
	})
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
