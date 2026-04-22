package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/handlers"
)

type MockSessionUseCase struct {
	mock.Mock
}

func (m *MockSessionUseCase) StartSession(ctx context.Context, sessionID uuid.UUID) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionUseCase) GetSessionDetails(ctx context.Context, sessionID uuid.UUID) (*domain.InterviewSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.InterviewSession), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestStartSession_InvalidID(t *testing.T) {
	mockUC := new(MockSessionUseCase)
	handler := handlers.NewSessionHandler(mockUC)

	r := httptest.NewRequest("POST", "/api/v1/sessions/invalid-uuid/start", nil)
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Post("/api/v1/sessions/{id}/start", handler.StartSession)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockUC.AssertNotCalled(t, "StartSession")
}

func TestStartSession_Success(t *testing.T) {
	mockUC := new(MockSessionUseCase)
	handler := handlers.NewSessionHandler(mockUC)
	sessionID := uuid.New()

	mockUC.On("StartSession", mock.Anything, sessionID).Return(nil)

	r := httptest.NewRequest("POST", "/api/v1/sessions/"+sessionID.String()+"/start", nil)
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Post("/api/v1/sessions/{id}/start", handler.StartSession)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "session started", response["message"])
}

func TestGetSession_NotFound(t *testing.T) {
	mockUC := new(MockSessionUseCase)
	handler := handlers.NewSessionHandler(mockUC)
	sessionID := uuid.New()

	mockUC.On("GetSessionDetails", mock.Anything, sessionID).Return(nil, errors.New("not found"))

	r := httptest.NewRequest("GET", "/api/v1/sessions/"+sessionID.String(), nil)
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Get("/api/v1/sessions/{id}", handler.GetSession)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
