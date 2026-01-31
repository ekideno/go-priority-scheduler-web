package web

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/ekideno/go-priority-scheduler-web/internal/scheduler"
)

type Handler struct {
	scheduler *scheduler.Scheduler
	templates *template.Template
}

func NewHandler(s *scheduler.Scheduler) *Handler {
	tmpl := template.Must(template.ParseFiles("internal/web/templates/index.html"))

	return &Handler{
		scheduler: s,
		templates: tmpl,
	}
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if err := h.templates.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		respondError(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Priority < 1 || req.Priority > 100 {
		respondError(w, "Priority must be between 1 and 100", http.StatusBadRequest)
		return
	}
	if req.Duration <= 0 {
		respondError(w, "Duration must be positive", http.StatusBadRequest)
		return
	}

	jobName := req.Name

	job := &scheduler.Job{
		Name:     jobName,
		Priority: req.Priority,
		Duration: time.Duration(req.Duration) * time.Second,
		Execute: func(ctx context.Context, d time.Duration) {
			select {
			case <-time.After(d):
				log.Printf("Job '%s' completed", jobName)
			case <-ctx.Done():
				log.Printf("Job '%s' cancelled", jobName)
			}
		},
	}

	h.scheduler.Schedule(job)

	respondJSON(w, CreateJobResponse{
		ID:      job.ID,
		Message: "Job scheduled successfully",
	}, http.StatusCreated)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	statuses := h.scheduler.GetStatus()
	respondJSON(w, statuses, http.StatusOK)
}

func (h *Handler) GetQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queue := h.scheduler.GetQueuedJobs()
	respondJSON(w, queue, http.StatusOK)
}

type CreateJobRequest struct {
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	Duration int    `json:"duration"`
}

type CreateJobResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, map[string]string{"error": message}, status)
}
