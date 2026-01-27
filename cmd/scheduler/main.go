package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ekideno/go-priority-scheduler-web/internal/web"
)

func main() {

	mux := http.NewServeMux()
	web.RegisterHandlers(mux)
	web.RegisterAPIRoutes(mux)

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		log.Println("Starting web server at http://localhost:8080")
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	s.Shutdown(ctxTimeout)

}
