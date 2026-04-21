package services

import (
	"context"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/username/app-recrutamento-ia/internal/domain"
	"github.com/username/app-recrutamento-ia/internal/logger"
	"go.uber.org/zap"
)

// InterviewOrchestrator gerencia o fluxo em tempo real da entrevista.
type InterviewOrchestrator struct {
	sttClient domain.STTClient
	llmClient domain.LLMClient
	ttsClient domain.TTSClient
	aiTrack   *lksdk.LocalSampleTrack
}

// NewInterviewOrchestrator constrói um novo orquestrador com as dependências de IA.
func NewInterviewOrchestrator(stt domain.STTClient, llm domain.LLMClient, tts domain.TTSClient) *InterviewOrchestrator {
	return &InterviewOrchestrator{
		sttClient: stt,
		llmClient: llm,
		ttsClient: tts,
	}
}

// SetupAITrack cria a trilha de áudio local para a IA e publica na sala.
func (o *InterviewOrchestrator) SetupAITrack(room *lksdk.Room) error {
	capability := webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: 48000,
		Channels:  1,
	}

	track, err := lksdk.NewLocalSampleTrack(capability)
	if err != nil {
		return err
	}

	_, err = room.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{
		Name:   "ia-voice",
		Source: livekit.TrackSource_MICROPHONE,
	})
	if err != nil {
		return err
	}

	o.aiTrack = track
	return nil
}

// HandleCandidateAudio processa o áudio vindo do candidato.
// Note que usamos *webrtc.TrackRemote pois o lksdk.RemoteTrack do LiveKit envolve essa interface para ler pacotes RTP.
func (o *InterviewOrchestrator) HandleCandidateAudio(ctx context.Context, track *webrtc.TrackRemote, rp *lksdk.RemoteParticipant) {
	audioStream := make(chan []byte, 100)
	defer close(audioStream)

	textStream, err := o.sttClient.TranscribeStream(ctx, audioStream)
	if err != nil {
		logger.Error("Failed to start STT stream", zap.Error(err))
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				rtpPacket, _, err := track.ReadRTP()
				if err != nil {
					return
				}
				audioStream <- rtpPacket.Payload
			}
		}
	}()

	for candidateText := range textStream {
		logger.Info("Candidate said:", zap.String("text", candidateText))
		go o.processLLMAndTTS(ctx, candidateText)
	}
}

func (o *InterviewOrchestrator) processLLMAndTTS(ctx context.Context, candidateText string) {
	systemPrompt := "Você é um recrutador técnico amigável. Faça perguntas curtas e diretas."
	history := []domain.SessionTurn{}

	llmStream, err := o.llmClient.GenerateResponseStream(ctx, systemPrompt, history, candidateText)
	if err != nil {
		logger.Error("Failed to generate LLM response", zap.Error(err))
		return
	}

	audioBytesStream, err := o.ttsClient.SynthesizeStream(ctx, llmStream)
	if err != nil {
		logger.Error("Failed to synthesize TTS", zap.Error(err))
		return
	}

	for audioChunk := range audioBytesStream {
		err = o.aiTrack.WriteSample(media.Sample{
			Data:     audioChunk,
			Duration: 20 * time.Millisecond,
		}, nil)
		
		if err != nil {
			logger.Error("Error writing sample to AI track", zap.Error(err))
		}
	}
}
