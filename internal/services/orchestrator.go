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

// Introduce faz a IA iniciar a conversa contextualizada com a vaga
func (o *InterviewOrchestrator) Introduce(ctx context.Context, jobTitle string) {
	// Atraso intencional para garantir que a conexão WebRTC no frontend estabilize
	time.Sleep(2 * time.Second)

	systemPrompt := "Você é a Nova, uma IA recrutadora gentil e objetiva."
	history := []domain.SessionTurn{}

	introPrompt := "Inicie a entrevista se apresentando brevemente e perguntando ao candidato se ele está pronto para a vaga de " + jobTitle + "."

	llmStream, err := o.llmClient.GenerateResponseStream(ctx, systemPrompt, history, introPrompt)
	if err != nil {
		logger.Error("Failed to generate intro", zap.Error(err))
		return
	}

	audioBytesStream, err := o.ttsClient.SynthesizeStream(ctx, llmStream)
	if err != nil {
		logger.Error("Failed to synthesize intro TTS", zap.Error(err))
		return
	}

	for audioChunk := range audioBytesStream {
		err = o.aiTrack.WriteSample(media.Sample{
			Data:     audioChunk,
			Duration: 20 * time.Millisecond,
		}, nil)

		if err != nil {
			logger.Error("Error writing intro sample", zap.Error(err))
		}
	}
}

// HandleCandidateAudio processa o áudio vindo do candidato.
func (o *InterviewOrchestrator) HandleCandidateAudio(ctx context.Context, track *webrtc.TrackRemote, rp *lksdk.RemoteParticipant) {
	audioStream := make(chan []byte, 100)
	defer close(audioStream)

	// Inicia a transcrição com Deepgram
	textStream, err := o.sttClient.TranscribeStream(ctx, audioStream)
	if err != nil {
		logger.Error("Failed to start STT stream", zap.Error(err))
		return
	}

	// Lê pacotes RTP vindos do Browser
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				rtpPacket, _, err := track.ReadRTP()
				if err != nil {
					logger.Error("Error reading RTP track (candidate left?)", zap.Error(err))
					return // Sai se o candidato mutou ou desconectou permanentemente
				}
				// IMPORTANTE: Em um cenário real de produção com Deepgram STT de áudio PCM,
				// precisaríamos usar um decoder Opus (pion/opus) para converter rtpPacket.Payload de OPUS para PCM
				// e depois mandar no canal. Por hora estamos injetando o payload Opus direto,
				// o que faz o Deepgram quebrar (Error: close 1011) se não estiver esperando OPUS puro.
				audioStream <- rtpPacket.Payload
			}
		}
	}()

	// Loop escutando os textos transcritos pelo Deepgram
	for {
		select {
		case <-ctx.Done():
			return
		case candidateText, ok := <-textStream:
			if !ok {
				return
			}
			logger.Info("Candidate said:", zap.String("text", candidateText))
			go o.processLLMAndTTS(ctx, candidateText)
		}
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
