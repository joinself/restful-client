package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	dbx "github.com/go-ozzo/ozzo-dbx"
	"github.com/joinself/restful-client/internal/account"
	"github.com/joinself/restful-client/internal/app"
	"github.com/joinself/restful-client/internal/attestation"
	"github.com/joinself/restful-client/internal/auth"
	"github.com/joinself/restful-client/internal/config"
	"github.com/joinself/restful-client/internal/connection"
	"github.com/joinself/restful-client/internal/entity"
	"github.com/joinself/restful-client/internal/fact"
	"github.com/joinself/restful-client/internal/healthcheck"
	"github.com/joinself/restful-client/internal/message"
	"github.com/joinself/restful-client/internal/notification"
	"github.com/joinself/restful-client/internal/request"
	"github.com/joinself/restful-client/internal/self"
	"github.com/joinself/restful-client/pkg/dbcontext"
	"github.com/joinself/restful-client/pkg/log"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/joinself/restful-client/docs" // docs is generated by Swag CLI, you have to import it.
	_ "github.com/lib/pq"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

func main() {
	flag.Parse()
	// create root logger tagged with server version
	logger := log.New().With(context.Background(), "version", Version)

	// load application configurations
	cfg, err := config.Load(logger, ".env")
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// connect to the database
	err = os.MkdirAll(cfg.StorageDir, 0744)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	db, err := dbx.MustOpen("sqlite3", filepath.Join(cfg.StorageDir, "client.db"))
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
//	@title			Joinself restful-client API
//	@version		1.0
//	@description	This is the api for Joinself restful client.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @host		localhost:8080
// @BasePath	/v1/
// @schemes		http https
func buildHandler(logger log.Logger, db *dbcontext.DB, cfg *config.Config) http.Handler {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	/*
		// TODO: move this to an environment variable.
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		}))
	*/
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	rg := e.Group("/v1")

	healthcheck.RegisterHandlers(rg, Version)

	authHandler := auth.Handler(cfg.JWTSigningKey)

	// Repositories
	connectionRepo := connection.NewRepository(db, logger)
	messageRepo := message.NewRepository(db, logger)
	factRepo := fact.NewRepository(db, logger)
	requestRepo := request.NewRepository(db, logger)
	attestationRepo := attestation.NewRepository(db, logger)
	accountRepo := account.NewRepository(db, logger)
	appRepo := app.NewRepository(db, logger)

	// Services
	rService := request.NewService(requestRepo, factRepo, attestationRepo, logger)
	runner := self.NewRunner(self.RunnerConfig{
		ConnectionRepo: connectionRepo,
		FactRepo:       factRepo,
		MessageRepo:    messageRepo,
		RequestRepo:    requestRepo,
		RequestService: rService,
		Logger:         logger,
		StorageKey:     cfg.StorageKey,
		StorageDir:     cfg.StorageDir,
	})
	rService.SetRunner(runner)
	cService := connection.NewService(connectionRepo, runner, logger)
	aService := app.NewService(appRepo, runner, logger)

	// Handlers
	app.RegisterHandlers(rg.Group(""),
		aService,
		authHandler,
		logger,
	)
	connection.RegisterHandlers(rg.Group(""),
		cService,
		authHandler,
		logger,
	)
	message.RegisterHandlers(rg.Group(""),
		message.NewService(messageRepo, runner, logger),
		cService,
		authHandler,
		logger,
	)
	fact.RegisterHandlers(rg.Group(""),
		fact.NewService(factRepo, attestationRepo, runner, logger),
		cService,
		authHandler, logger,
	)
	request.RegisterHandlers(rg.Group(""),
		rService,
		cService,
		authHandler, logger,
	)
	auth.RegisterHandlers(rg.Group(""),
		auth.NewService(cfg, accountRepo, logger),
		logger,
	)
	account.RegisterHandlers(rg.Group(""),
		account.NewService(accountRepo, logger),
		authHandler,
		logger,
	)
	notification.RegisterHandlers(rg.Group(""),
		notification.NewService(runner, logger),
		authHandler, logger,
	)

	if cfg.DefaultSelfApp != nil {
		runner.Run(entity.App{
			ID:           cfg.DefaultSelfApp.SelfAppID,
			DeviceSecret: cfg.DefaultSelfApp.SelfAppDeviceSecret,
			Env:          cfg.DefaultSelfApp.SelfEnv,
			Callback:     cfg.DefaultSelfApp.CallbackURL,
		})
	}
	/*
		for id, client := range clients {
			logger.Infof("starting client %s", id)
			self.RunService(
				self.NewService(self.Config{
					SelfClient:     support.NewSelfClient(client),
					ConnectionRepo: connectionRepo,
					FactRepo:       factRepo,
					MessageRepo:    messageRepo,
					RequestRepo:    requestRepo,
					RequestService: rService,
					Logger:         logger,
					Poster:         webhook.NewWebhook(callbackURLs[id]),
				}),
				logger,
			)
		}
	*/
	if cfg.ServeDocs == "true" {
		e.GET("/docs/*", echoSwagger.WrapHandler)
	}

	// Start server
	fmt.Println(e.Start(":" + strconv.Itoa(cfg.ServerPort)))

	return e
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
