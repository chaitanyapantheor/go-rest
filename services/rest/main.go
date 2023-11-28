package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/chaitanyamaili/go_rest/pkg/database"
	"github.com/chaitanyamaili/go_rest/pkg/logger"
	"github.com/chaitanyamaili/go_rest/services/rest/handlers"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql"
)

const (
	appName = "rest"
)

var (
	appVersionLDFlag        string
	appBuildTimestampLDFlag string
)

func main() {
	// -------------------------------------------------------------------
	// Logger
	// -------------------------------------------------------------------
	log, err := logger.GetProductionLogger(appName, appVersionLDFlag)
	if err != nil {
		fmt.Println("error initializing production logger")
		os.Exit(1)
	}

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup failure", "ERROR", err)

		_ = log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	var err error
	viperConfigJSONFile := "rest.local"
	viper.SetConfigName(viperConfigJSONFile)
	viper.SetConfigType("json")
	// Reads cloud function configuration in GCP as a mounted secrets.
	// Same path should be used while provisioning the secret to CF.
	viper.AddConfigPath("/functions/config")
	// Reads local system configurations.
	viper.AddConfigPath(".")
	// Find and read the config file
	err = viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// -------------------------------------------------------------------
	// Startup Details
	// -------------------------------------------------------------------
	log.Infow("startup", "binary build time", appBuildTimestampLDFlag)

	// -------------------------------------------------------------------
	// Databases
	// -------------------------------------------------------------------
	log.Infow("startup.db", "status", "initializing DBs")

	db, err := database.Open(database.Config{
		Type:         viper.GetString("db.type"),
		User:         viper.GetString("db.user"),
		Password:     viper.GetString("db.password"),
		Host:         viper.GetString("db.host"),
		Port:         viper.GetInt("db.port"),
		Name:         viper.GetString("db.dbName"),
		MaxIdleConns: viper.GetInt("db.maxIdleConns"),
		MaxOpenConns: viper.GetInt("db.maxOpenConns"),
		DisableTLS:   viper.GetBool("db.disableTLS"),
	})
	if err != nil {
		panic(fmt.Errorf("connecting to db: %w", err))
	}
	defer func() {
		log.Infow("shutdown", "status", "stopping db", "host", viper.GetString("db.host"))
		_ = db.Close()
	}()

	// -------------------------------------------------------------------
	// Initialize API
	// -------------------------------------------------------------------
	log.Infow("startup.api", "status", "initializing API")

	// -------------------------------------------------------------------
	// RWMux for lock DBs in transaction mode (deadlocks = yuck)
	// -------------------------------------------------------------------
	log.Infow("startup.remux", "status", "created")
	rwmux := &sync.RWMutex{}

	apiMux := handlers.APIMux(handlers.APIMuxConfig{
		Log:     log,
		DB:      db,
		RWMux:   rwmux,
		Headers: viper.GetBool("app.enforceHeaders"),
	})

	// -------------------------------------------------------------------
	// New Channels
	// -------------------------------------------------------------------

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	subscribeErrors := make(chan error, 1)

	apiHost := fmt.Sprintf("%s:%s", viper.GetString("web.apiHost"), viper.GetString("web.apiPort"))
	api := http.Server{
		Addr:              apiHost,
		Handler:           apiMux,
		ReadHeaderTimeout: viper.GetDuration("web.readHeaderTimeout"),
		ReadTimeout:       viper.GetDuration("web.readTimeout"),
		WriteTimeout:      viper.GetDuration("web.writeTimeout"),
		IdleTimeout:       viper.GetDuration("web.idleTimeout"),
		MaxHeaderBytes:    viper.GetInt("web.maxHeaderBytes"),
	}

	// -------------------------------------------------------------------
	// Starting the API
	// -------------------------------------------------------------------

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup.api", "status", "api router started", "host", api.Addr)

		log.Debugf("\n%s\nAPI STARTED\nHost: %s\n%s\n",
			strings.Repeat("-", 75),
			api.Addr,
			strings.Repeat("-", 75),
		)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------
	// Shutdown
	// -------------------------------------------------------------------

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case err := <-subscribeErrors:
		return fmt.Errorf("subscriber error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown completed", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("web.shutdownTimeout"))
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := api.Shutdown(ctx); err != nil {
			_ = api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
