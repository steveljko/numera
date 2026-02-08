package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"numera/config"
	"numera/db"
	"numera/handler"
	"numera/pkg/session"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

const defaultShutdownPeriod = 30 * time.Second

type App struct {
	addr    string
	cfg     *config.Config
	db      *sql.DB
	logger  *logrus.Logger
	session *session.Session
	wg      sync.WaitGroup
}

func NewApp(
	addr string,
	cfg *config.Config,
	db *sql.DB,
	logger *logrus.Logger,
	session *session.Session,
) *App {
	return &App{
		addr:    addr,
		cfg:     cfg,
		db:      db,
		logger:  logger,
		session: session,
	}
}

func (app *App) Serve() error {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger:  app.logger,
		NoColor: false,
	}))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   app.cfg.AllowedOrigins,
		AllowedMethods:   app.cfg.AllowedMethods,
		AllowedHeaders:   app.cfg.AllowedHeaders,
		ExposedHeaders:   app.cfg.ExposedHeaders,
		AllowCredentials: app.cfg.AllowCredentials,
		MaxAge:           app.cfg.MaxAge,
	}))
	r.Use(app.session.LoadAndSave)

	// serve static files
	fs := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	userHandler := handler.NewUserHandler(app.db, app.logger, app.session)
	userHandler.RegisterRoutes(r)

	authHandler := handler.NewAuthHandler(app.db, app.logger, app.session)
	authHandler.RegisterRoutes(r)

	dashboardHandler := handler.NewDashboardHandler(app.db, app.logger, app.session)
	dashboardHandler.RegisterRoutes(r)

	accountHandler := handler.NewAccountHandler(app.db, app.logger, app.session)
	accountHandler.RegisterRoutes(r)

	if !app.cfg.IsProd() {
		printRoutes(r, app.logger)
	}

	server := &http.Server{
		Addr:    app.addr,
		Handler: r,
	}

	// graceful shutdown handler
	shutdownErrorChan := make(chan error, 1)
	go func() {
		quitChan := make(chan os.Signal, 1)
		signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-quitChan

		app.logger.Info("shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownPeriod)
		defer cancel()

		shutdownErrorChan <- server.Shutdown(ctx)
	}()

	app.logger.WithFields(logrus.Fields{
		"addr": server.Addr,
		"mode": app.cfg.Mode,
	}).Info("starting server")

	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// wait for shutdown to complete
	err = <-shutdownErrorChan
	if err != nil {
		app.logger.WithError(err).Error("error during server shutdown")
		return err
	}

	if app.db != nil {
		app.logger.Info("closing database connection")
		if err := app.db.Close(); err != nil {
			app.logger.WithError(err).Error("error closing database")
			return err
		}
	}

	app.logger.Info("waiting for background tasks to complete")
	app.wg.Wait()

	app.logger.WithField("addr", server.Addr).Info("server stopped gracefully")
	return nil
}

func setupLogger(mode string) *logrus.Logger {
	logger := logrus.New()

	if mode == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.SetOutput(os.Stdout)

	return logger
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("failed to load config")
	}

	logger := setupLogger(cfg.Mode)

	dbConn, err := db.Open(cfg.DBPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to open database")
	}

	session := session.New(dbConn, cfg)
	// if err != nil {
	// 	logger.WithError(err).Fatal("failed to initialize session")
	// }

	const migrationDir = "migrations/"
	logger.Info("running database migrations")
	if err := db.RunMigrations(dbConn, migrationDir); err != nil {
		logger.WithError(err).Fatal("failed to run migrations")
	}

	if cfg.Port == "" {
		logger.Fatal("port is not provided")
	}

	server := NewApp(fmt.Sprintf(":%s", cfg.Port), cfg, dbConn, logger, session)
	if err := server.Serve(); err != nil {
		logger.WithError(err).Fatal("server failed")
	}
}

func printRoutes(r *chi.Mux, logger *logrus.Logger) {
	logger.Info("=== REGISTERED ROUTES ===")
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		logger.WithFields(logrus.Fields{
			"method": method,
			"route":  route,
		}).Info("route")
		return nil
	}
	if err := chi.Walk(r, walkFunc); err != nil {
		logger.WithError(err).Error("Failed to walk routes")
	}
	logger.Info("=== END ROUTES ===")
}
