package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/livekit"
	"github.com/username/app-recrutamento-ia/templates/pages"
)

type FrontendHandler struct{}

func NewFrontendHandler() *FrontendHandler {
	return &FrontendHandler{}
}

// ServeDashboardPage renders the main dashboard for HR
func (h *FrontendHandler) ServeDashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// TODO: Fetch real metrics from the database
	data := domain.DashboardData{
		Metrics: domain.DashboardMetrics{
			TotalCandidates: 120,
			ActiveSessions:  2,
			DoneSessions:    118,
		},
		RecentInterviews: []domain.RecentInterview{
			{CandidateName: "João Silva", JobTitle: "Desenvolvedor Go", Status: domain.SessionStatusDone, Score: "8.5/10"},
			{CandidateName: "Maria Oliveira", JobTitle: "Engenheira de Dados", Status: domain.SessionStatusInProgress, Score: "-"},
		},
	}

	component := pages.DashboardHome(data)
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// ServeVagasPage renders the jobs management dashboard page
func (h *FrontendHandler) ServeVagasPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Placeholder data for design
	component := pages.VagasHome()
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
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
