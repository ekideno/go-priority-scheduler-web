package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	web "github.com/ekideno/go-priority-scheduler-web/internal/web/handlers"
	"github.com/ekideno/go-priority-scheduler-web/internal/web/scheduler"
)

func main() {

	sched := scheduler.New(5)
	router := web.NewRouter(sched)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		log.Println("Starting web server at http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelTimeout()
	srv.Shutdown(ctxTimeout)
	sched.Shutdown()

}
