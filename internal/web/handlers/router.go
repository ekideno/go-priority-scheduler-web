package web

import (
	"log"
	"net/http"
	"time"

	"github.com/ekideno/go-priority-scheduler-web/internal/web/scheduler"
)

func NewRouter(s *scheduler.Scheduler) http.Handler {
	handler := NewHandler(s)
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.Home)

	mux.HandleFunc("/api/jobs", handler.CreateJob)
	mux.HandleFunc("/api/status", handler.GetStatus)
	mux.HandleFunc("/api/queue", handler.GetQueue)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	return loggingMiddleware(corsMiddleware(mux))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
