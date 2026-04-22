package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"

	"github.com/username/app-recrutamento-ia/internal/infrastructure/deepgram"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/elevenlabs"
	"github.com/username/app-recrutamento-ia/internal/infrastructure/openai"
	zlog "github.com/username/app-recrutamento-ia/internal/logger"
	"github.com/username/app-recrutamento-ia/internal/services"
	"go.uber.org/zap"
)

func main() {
	_ = godotenv.Load()
	zlog.InitLogger()
	defer zlog.Get().Sync()

	hostURL := getEnvOrDefault("LIVEKIT_URL", "ws://127.0.0.1:7880")
	apiKey := getEnvOrDefault("LIVEKIT_API_KEY", "devkey")
	apiSecret := getEnvOrDefault("LIVEKIT_API_SECRET", "secret")

	// Inicializa Clientes de IA
	sttClient, err := deepgram.NewSTTClient(getEnvOrDefault("DEEPGRAM_API_KEY", "dummy"))
	if err != nil {
		zlog.Fatal("Failed to init STT")
	}

	llmClient, err := openai.NewLLMClient(getEnvOrDefault("OPENAI_API_KEY", "dummy"))
	if err != nil {
		zlog.Fatal("Failed to init LLM")
	}

	ttsClient, err := elevenlabs.NewTTSClient(
		getEnvOrDefault("ELEVENLABS_API_KEY", "dummy"),
		getEnvOrDefault("ELEVENLABS_VOICE_ID", "dummy"),
	)
	if err != nil {
		zlog.Fatal("Failed to init TTS")
	}

	// Orquestrador da Entrevista
	orchestrator := services.NewInterviewOrchestrator(sttClient, llmClient, ttsClient)

	log.Printf("Conectando ao LiveKit Server: %s", hostURL)

	roomName := "interview-test"
	room, err := lksdk.ConnectToRoom(hostURL, lksdk.ConnectInfo{
		APIKey:              apiKey,
		APISecret:           apiSecret,
		RoomName:            roomName,
		ParticipantIdentity: "ai-worker",
	}, &lksdk.RoomCallback{
		ParticipantCallback: lksdk.ParticipantCallback{
			OnTrackSubscribed: func(track *webrtc.TrackRemote, pub *lksdk.RemoteTrackPublication, rp *lksdk.RemoteParticipant) {
				zlog.Info("Track Subscribed", zap.String("track_id", track.ID()))
				if track.Kind() == webrtc.RTPCodecTypeAudio {
					go orchestrator.HandleCandidateAudio(context.Background(), track, rp)
				}
			},
		},
		OnParticipantConnected: func(p *lksdk.RemoteParticipant) {
			log.Printf("Participante conectado: %s", p.Identity())
		},
	})

	if err != nil {
		zlog.Error("Falha ao conectar ao LiveKit", zap.Error(err))
		os.Exit(1)
	}

	// Prepara a trilha de áudio da IA para falar
	err = orchestrator.SetupAITrack(room)
	if err != nil {
		zlog.Fatal("Failed to setup AI audio track")
	}

	log.Printf("Conectado à sala %s com sucesso!", room.Name())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Desconectando do LiveKit...")
	room.Disconnect()
	log.Println("Worker encerrado.")
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
