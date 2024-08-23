package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/guerzon/gobudget-api/pkg/api"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/guerzon/gobudget-api/pkg/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

//	@title			gobudget API
//	@version		beta
//	@description	gobudget API

//	@host		localhost:8080
//	@BasePath	/beta
//	@accept		json
//	@produce	json

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/

func main() {

	// Load config
	dir, _ := os.Getwd()
	config, err := util.LoadConfig(dir)
	if err != nil {
		slog.Error("startup failed", "action", "load configuration", "errmsg", err)
		os.Exit(1)
	}

	// Create a DB connection
	conn, err := pgxpool.New(context.Background(), config.DBConnString)
	if err != nil {
		slog.Error("startup failed", "action", "database connection", "error", err)
		os.Exit(2)
	}
	dbStore := db.NewStore(conn)

	// Create a Redis client
	redisOpts := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpts)

	// Create an email sender
	var emailSender util.EmailSender
	if config.Environment == "local" {
		emailSender = util.NewLocalSender(config.EmailSenderName, config.MailhogSenderAddress, config.MailhogHost)

	} else {
		emailSender = util.NewGmailSender(config.EmailSenderName, config.GmailSenderAddress, config.GmailSenderPassword)
	}

	// Run the db migration
	dbMigration(config.DBMigrationFiles, config.DBConnString)

	// Create the server
	apiServer, err := api.NewServer(config, dbStore, taskDistributor)
	if err != nil {
		slog.Error("startup failed", "action", "create server instance", "errmsg", err)
		os.Exit(5)
	}

	// Start the task processor in a go routine
	go startTaskProcessor(redisOpts, dbStore, emailSender)

	// Use http from the std library
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", config.ListenAddr, config.ListenPort),
		Handler:           apiServer.Router,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Setup routine to shut down the server gracefully
	idleConnsClosed := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
		<-s

		slog.Info("interrupt received ...")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("cannot shut down server: %s", err)
		}
		close(idleConnsClosed)
	}()

	// Start the server
	if err = srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("startup failed", "action", "start API server", "errmsg", err)
		}
	}

	<-idleConnsClosed
	slog.Info("API server has been shut down.")
}

func dbMigration(migrationFiles string, dbConnString string) {

	slog.Info("starting db migration")

	mig, err := migrate.New(migrationFiles, dbConnString)
	if err != nil {
		slog.Error("startup failed", "action", "create db migration", "error", err)
		os.Exit(2)
	}

	if err = mig.Up(); err != nil {
		if err != migrate.ErrNoChange {
			slog.Error("startup failed", "action", "migrate database", "error", err)
			os.Exit(3)
		}
		slog.Info("no change detected during db migration")
	}

	slog.Info("db migration completed successfully")
}

// Starts the task processor for picking up tasks from Redis. It receives a db.Store and a util.EmailSender object for any DB and email task it requires.
func startTaskProcessor(redisOpts asynq.RedisClientOpt, store db.Store, mailer util.EmailSender) {

	processor := worker.NewRedisTaskProcessor(redisOpts, store, mailer)
	slog.Info("Starting task processor ...")

	err := processor.Start()
	if err != nil {
		slog.Error("startup failed", "action", "start task processor", "error", err)
		os.Exit(7)
	}
	slog.Info("Task processor has started.")
}
