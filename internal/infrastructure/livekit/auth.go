package livekit

import (
	"fmt"
	"os"
	"time"

	"github.com/livekit/protocol/auth"
)

// TokenConfig holds configuration for generating a token.
type TokenConfig struct {
	APIKey    string
	APISecret string
	RoomName  string
	Identity  string
	Name      string
	Duration  time.Duration
}

// GenerateCandidateToken creates a secure LiveKit token for the candidate to join a specific room.
func GenerateCandidateToken(roomName, candidateID, candidateName string) (string, error) {
	apiKey := getEnvOrDefault("LIVEKIT_API_KEY", "devkey")
	apiSecret := getEnvOrDefault("LIVEKIT_API_SECRET", "super-secret-key-that-must-be-32-chars")

	// Create a new AccessToken
	at := auth.NewAccessToken(apiKey, apiSecret)

	// Set identity and name
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}

	// Apply the grant
	at.AddGrant(grant).
		SetIdentity(candidateID).
		SetName(candidateName).
		SetValidFor(2 * time.Hour) // The token is valid for 2 hours

	// Sign the token
	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("failed to generate LiveKit token: %w", err)
	}

	return token, nil
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
