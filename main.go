package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/eco_codes/handler"
	"github.com/eco_codes/scraper"
	"github.com/gorilla/mux"
)

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	ecoCodeDataURL := os.Getenv("ECO_CODE_DATA_URL")
	if strings.Trim(ecoCodeDataURL, " ") == "" {
		ecoCodeDataURL = "https://www.chessgames.com/chessecohelp.html"
	}

	ws := scraper.NewWebScraper(ecoCodeDataURL)

	ch := handler.NewChessMoveHandler(ws)

	r := mux.NewRouter()

	r.HandleFunc("/", ch.Get).Methods(http.MethodGet)
	r.HandleFunc("/{CODE}", ch.GetMoves).Methods(http.MethodGet)
	r.HandleFunc("/next/{MOVE:.*}", ch.GetNextMove).Methods(http.MethodGet) // Route for chess player next move

	port := os.Getenv("PORT")
	addr := fmt.Sprintf("0.0.0.0:%s", port)

	srv := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 120,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
