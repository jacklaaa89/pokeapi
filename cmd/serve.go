package cmd

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/jacklaaa89/pokeapi/internal/api/log/fmt"
	"github.com/jacklaaa89/pokeapi/internal/server"
	"github.com/jacklaaa89/pokeapi/internal/server/middleware"
)

// defaultPort the default port to bind to when --port is not supplied.
const defaultPort = 5555

// port is the defined port binding.
var port int

// logLevel is the level of logging to apply.
var logLevel int

// serveCmd this is the command which initialises and starts the HTTP API
// --port can be used to override the port in which to bind to.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serves the API over HTTP",
	Long:  "Serves the PokeAPI over HTTP",
	RunE:  serve,
}

// init sets up the persistent flag bindings.
func init() {
	serveCmd.PersistentFlags().IntVar(&port, "port", defaultPort, "the port to listen on")
	serveCmd.PersistentFlags().IntVar(
		&logLevel, "log-level", int(fmt.LevelError),
		"the logging level, can be one of 0 (None), 1 (Error), 2 (Warn), 3 (Info), 4 (Debug)",
	)
}

// serve initialises a HTTP server and listens for any incoming connections
// this function will block until either the server is shutdown from elsewhere or
// a signal is received.
func serve(cmd *cobra.Command, _ []string) error {
	log.SetOutput(cmd.OutOrStdout())
	addr := ":" + strconv.Itoa(port)

	log.Printf("listening on port: %d\n", port)

	l := fmt.New(fmt.Level(logLevel))
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	svr := &http.Server{
		Addr: addr,
		Handler: server.Handler(
			middleware.WithRequestID(),
			middleware.WithLogger(l),
		),
	}

	// listen in a new go-routine so we can handle signal interrupts etc.
	go func() {
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// listen for any interrupts
	<-sig
	close(sig)

	log.Println("\nsignal received - shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := svr.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %+v", err)
	}

	return nil
}
