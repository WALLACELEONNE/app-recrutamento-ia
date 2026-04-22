package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/livekit"
	"github.com/username/app-recrutamento-ia/templates/pages"
)

type FrontendHandler struct {
	sessionRepo domain.SessionRepository
}

func NewFrontendHandler(sessionRepo domain.SessionRepository) *FrontendHandler {
	return &FrontendHandler{
		sessionRepo: sessionRepo,
	}
}

// ServeDashboardPage renders the main dashboard for HR
func (h *FrontendHandler) ServeDashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	ctx := r.Context()

	metrics, err := h.sessionRepo.GetDashboardMetrics(ctx)
	if err != nil {
		metrics = domain.DashboardMetrics{} // fallback or handle error
	}

	interviews, err := h.sessionRepo.GetRecentInterviews(ctx, 10)
	if err != nil {
		interviews = []domain.RecentInterview{} // fallback or handle error
	}

	data := domain.DashboardData{
		Metrics:          metrics,
		RecentInterviews: interviews,
	}

	component := pages.DashboardHome(data)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// ServeVagasPage renders the jobs management dashboard page
func (h *FrontendHandler) ServeVagasPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	jobs, err := h.sessionRepo.GetJobs(r.Context())
	if err != nil {
		jobs = []domain.Job{}
	}

	component := pages.VagasHome(jobs)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// HandleCreateVaga creates a new job from the form
func (h *FrontendHandler) HandleCreateVaga(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	department := r.FormValue("department")

	if title == "" || department == "" {
		http.Error(w, "Title and department are required", http.StatusBadRequest)
		return
	}

	job := &domain.Job{
		ID:         uuid.New(),
		Title:      title,
		Department: department,
		CreatedAt:  time.Now(),
	}

	err = h.sessionRepo.CreateJob(r.Context(), job)
	if err != nil {
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/vagas", http.StatusSeeOther)
}

// ServeInterviewPage renders the interview UI for a candidate
func (h *FrontendHandler) ServeInterviewPage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	// In a real scenario, you'd validate the sessionID exists in DB and is active
	// For now, we generate a secure token for the candidate using the LiveKit auth package
	token, err := livekit.GenerateCandidateToken("room-"+sessionID, "candidate-"+sessionID, "Candidato")
	if err != nil {
		http.Error(w, "Error generating session token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the Templ component
	component := pages.InterviewRoom(sessionID, token)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
