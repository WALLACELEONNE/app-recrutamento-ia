package main

import (
	"context"
	"encoding/json"
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
	"github.com/username/app-recrutamento-ia/internal/infrastructure/queue"
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
		getEnvOrDefault("ELEVENLABS_VOICE_ID", "21m00Tcm4TlvDq8ikWAM"), // Default to Rachel if not provided
	)
	if err != nil {
		zlog.Fatal("Failed to init TTS")
	}

	redisURL := getEnvOrDefault("REDIS_URL", "127.0.0.1:6379")
	redisQueue, err := queue.NewRedisQueue(redisURL, "interview_jobs")
	if err != nil {
		zlog.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisQueue.Close()

	log.Printf("Worker iniciado e aguardando entrevistas via Redis (%s)", redisURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go redisQueue.Listen(ctx, func(ctx context.Context, jobData []byte) error {
		var job map[string]string
		if err := json.Unmarshal(jobData, &job); err != nil {
			return err
		}

		roomName := job["room_name"]
		if roomName == "" {
			return nil
		}

		zlog.Info("Iniciando AI Worker para a sala", zap.String("room_name", roomName))

		orchestrator := services.NewInterviewOrchestrator(sttClient, llmClient, ttsClient)

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
						// IMPORTANTE: precisamos aguardar ou capturar o payload em background de maneira resiliente
						go orchestrator.HandleCandidateAudio(context.Background(), track, rp)
					}
				},
			},
			OnParticipantConnected: func(p *lksdk.RemoteParticipant) {
				log.Printf("Participante conectado na sala %s: %s", roomName, p.Identity())
			},
			OnDisconnected: func() {
				log.Printf("Worker desconectado da sala %s", roomName)
			},
		})

		if err != nil {
			zlog.Error("Falha ao conectar ao LiveKit", zap.Error(err))
			return nil // Retornar nil para o redisQueue.Listen tentar continuar lendo (ou tratar de acordo)
		}

		err = orchestrator.SetupAITrack(room)
		if err != nil {
			zlog.Error("Failed to setup AI audio track", zap.Error(err))
			room.Disconnect()
			return err
		}

		// A IA se introduz para iniciar a entrevista
		go orchestrator.Introduce(ctx, "Desenvolvedor de Software") // Pode vir do payload do job futuramente

		log.Printf("Conectado à sala %s com sucesso!", room.Name())
		return nil
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Worker encerrado.")
}

func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
