package handlers

import (
	"net/http"
	"os"
	"strconv"
	"strings"
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

	jobs, err := h.sessionRepo.GetJobs(ctx)
	if err != nil {
		jobs = []domain.Job{}
	}

	data := domain.DashboardData{
		Metrics:          metrics,
		RecentInterviews: interviews,
		AvailableJobs:    jobs,
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

// ServeJobDetailsPage renders the specific job page
func (h *FrontendHandler) ServeJobDetailsPage(w http.ResponseWriter, r *http.Request) {
	jobIDStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.sessionRepo.GetJobByID(r.Context(), jobID)
	if err != nil {
		http.Error(w, "Failed to load job", http.StatusInternalServerError)
		return
	}

	interviews, err := h.sessionRepo.GetInterviewsByJobID(r.Context(), jobID)
	if err != nil {
		interviews = []domain.RecentInterview{}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := pages.VagaDetails(*job, interviews)
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
	nQuestions, _ := strconv.Atoi(r.FormValue("n_questions"))
	maxMinutes, _ := strconv.Atoi(r.FormValue("max_minutes"))
	persona := r.FormValue("persona")
	tone := r.FormValue("tone")

	focusAreasStr := r.FormValue("focus_areas")
	var focusAreas []string
	if focusAreasStr != "" {
		for _, area := range strings.Split(focusAreasStr, ",") {
			focusAreas = append(focusAreas, strings.TrimSpace(area))
		}
	}

	if title == "" || department == "" {
		http.Error(w, "Title and department are required", http.StatusBadRequest)
		return
	}

	if nQuestions == 0 {
		nQuestions = 5
	}
	if maxMinutes == 0 {
		maxMinutes = 15
	}

	job := &domain.Job{
		ID:         uuid.New(),
		Title:      title,
		Department: department,
		InterviewConfig: domain.JobConfig{
			NQuestions: nQuestions,
			MaxMinutes: maxMinutes,
			Persona:    persona,
			Tone:       tone,
			FocusAreas: focusAreas,
		},
		CreatedAt: time.Now(),
	}

	err = h.sessionRepo.CreateJob(r.Context(), job)
	if err != nil {
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard/vagas", http.StatusSeeOther)
}

// HandleInviteCandidate handles the creation of a new candidate and interview session
func (h *FrontendHandler) HandleInviteCandidate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	jobIDStr := r.FormValue("job_id")

	if name == "" || email == "" || jobIDStr == "" {
		http.Error(w, "Name, email and job are required", http.StatusBadRequest)
		return
	}

	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	candidate := &domain.Candidate{
		ID:    uuid.New(),
		Name:  name,
		Email: email,
		JobID: jobID,
	}

	err = h.sessionRepo.CreateCandidate(r.Context(), candidate)
	if err != nil {
		http.Error(w, "Failed to create candidate", http.StatusInternalServerError)
		return
	}

	session := &domain.InterviewSession{
		ID:          uuid.New(),
		CandidateID: candidate.ID,
		JobID:       jobID,
		Status:      domain.SessionStatusInvited,
	}

	err = h.sessionRepo.Create(r.Context(), session)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// ServeReportPage renders the detailed interview report
func (h *FrontendHandler) ServeReportPage(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	reportData, err := h.sessionRepo.GetSessionReport(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Failed to load report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	component := pages.InterviewReport(reportData)
	err = component.Render(r.Context(), w)
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

	livekitURL := os.Getenv("LIVEKIT_URL")
	if livekitURL == "" {
		livekitURL = "ws://127.0.0.1:7880"
	}

	// Trigger the AI Worker to join this room via Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "127.0.0.1:6379"
	}
	redisQ, err := queue.NewRedisQueue(redisURL, "interview_jobs")
	if err == nil {
		defer redisQ.Close()
		jobData := map[string]string{"room_name": "room-" + sessionID}
		_ = redisQ.Enqueue(r.Context(), jobData)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Render the Templ component
	component := pages.InterviewRoom(sessionID, token, livekitURL)
	err = component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
