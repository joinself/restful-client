package main

import (
	"context"
	"database/sql"
	"flag"
	lg "log"
	"net/http"
	"os"
	"strconv"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/healthcheck"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/internal/self"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
	selfsdk "github.com/joinself/self-go-sdk"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	buildHandler(logger, dbcontext.New(db), cfg)
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB, cfg *config.Config) http.Handler {
	client, err := setupSelfClient(cfg)
	if err != nil {
		lg.Fatalf("failed to setup self client: %v", err.Error())
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	healthcheck.RegisterHandlers(e, Version)
	rg := e.Group("/v1")
	// rg := router.Group("/v1")

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
		auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, cfg.User, cfg.Password, logger),
		logger,
	)

	self.RegisterHandlers(
		self.NewService(client, connectionRepo, factRepo, messageRepo, logger),
		logger,
	)

	// Start server
	e.Logger.Fatal(e.Start(":" + strconv.Itoa(cfg.ServerPort)))

	return e
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
