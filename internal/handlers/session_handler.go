package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"github.com/username/app-recrutamento-ia/internal/usecase"
	"go.uber.org/zap"
)

type SessionHandler struct {
	uc usecase.SessionUseCase
}

func NewSessionHandler(uc usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{
		uc: uc,
	}
}

// StartSession godoc
// @Summary Starts an interview session
// @Description Changes the status of an interview session to in_progress
// @Tags session
// @Accept json
// @Produce json
// @Param id path string true "Session UUID"
// @Success 200 {object} map[string]string "{"message": "session started"}"
// @Failure 400 {object} map[string]string "{"error": "invalid session id"}"
// @Failure 500 {object} map[string]string "{"error": "internal server error"}"
// @Router /api/v1/sessions/{id}/start [post]
func (h *SessionHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	err = h.uc.StartSession(r.Context(), sessionID)
	if err != nil {
		logger.Error("Error starting session", zap.Error(err))
		if err.Error() == "session is already done" {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondWithError(w, http.StatusInternalServerError, "failed to start session")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "session started"})
}

// GetSession godoc
// @Summary Gets interview session details
// @Description Retrieves metadata about an interview session
// @Tags session
// @Produce json
// @Param id path string true "Session UUID"
// @Success 200 {object} domain.InterviewSession
// @Failure 400 {object} map[string]string "{"error": "invalid session id"}"
// @Failure 404 {object} map[string]string "{"error": "session not found"}"
// @Router /api/v1/sessions/{id} [get]
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idParam)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	session, err := h.uc.GetSessionDetails(r.Context(), sessionID)
	if err != nil {
		logger.Error("Error fetching session", zap.Error(err))
		h.respondWithError(w, http.StatusNotFound, "session not found")
		return
	}

	h.respondWithJSON(w, http.StatusOK, session)
}

func (h *SessionHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, map[string]string{"error": message})
}

func (h *SessionHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
