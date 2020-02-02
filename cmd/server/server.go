package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"cryptic-command/gatewatch/handlers"
)

const (
	terminationTimeout = 5 * time.Second
)

var (
	port = flag.Uint("port", 3000, "tcp port number to listen and serve.")
	host = flag.String("host", "localhost:3000", "hostname for making interaction url")
)

func routing() http.Handler {
	r := mux.NewRouter()

	// debug handler
	r.Handle("/transaction", &handlers.TransactionHandler{InteractionHost: *host})
	r.Handle("/interact/{handle}", &handlers.InteractionHandler{})
	return r
}
func main() {
	log.Println("[INFO] xyz as server start")
	// parse commandline options
	flag.Parse()

	var (
		serverErrChan = make(chan error, 1)
	)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: routing(),
	}
	go func() {
		err := srv.ListenAndServe()
		serverErrChan <- err
	}()

	log.Println("[INFO] start listening...")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sigCh:
		log.Println("[INFO] termination signal received")
	case err := <-serverErrChan:
		if err != nil {
			log.Printf("[ERROR] server stopped abnormally: %v", err)
		}
	}

	tctx, tcancel := context.WithTimeout(context.Background(), terminationTimeout)
	defer tcancel()
	if err := srv.Shutdown(tctx); err != nil {
		log.Printf("[ERROR] server teardown error: %v", err)
	} else {
		log.Println("[INFO] stop listening...")
	}

	log.Println("[INFO] xyz as server end")
}
