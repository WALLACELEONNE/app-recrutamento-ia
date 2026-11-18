# Recrutamento com IA por voz

SaaS multi-tenant que conduz **entrevistas automatizadas por voz, em tempo real**. O candidato conversa com uma IA pelo navegador; a plataforma transcreve, gera a resposta e fala de volta em streaming, com meta de latência abaixo de 1,2 s.

Escrito inteiramente em **Go**, com Clean Architecture.

## Como funciona

```
Candidato (navegador) ──WebRTC──▶ LiveKit ──▶ AI Worker
                                                │
        Deepgram (STT) ◀── áudio ───────────────┤
        OpenAI GPT-4o  ◀── texto + histórico ───┤
        ElevenLabs (TTS) ◀── tokens em stream ──┘
                                                │
Candidato ◀── áudio da IA ◀── LiveKit ◀─────────┘
```

1. O candidato fala; o LiveKit entrega o áudio ao Worker.
2. O Worker envia o áudio ao Deepgram (STT) por WebSocket.
3. Com a transcrição final, o texto e o histórico vão à OpenAI (GPT-4o), que responde em stream.
4. Os tokens seguem incrementalmente para a ElevenLabs (TTS).
5. O áudio gerado volta ao LiveKit e toca no fone do candidato.

## Arquitetura

| Pasta | Responsabilidade |
|---|---|
| `cmd/server` | Servidor HTTP e painel web (RH e setup do candidato) |
| `cmd/worker` | Worker de IA que entra na sala WebRTC e orquestra STT, LLM e TTS |
| `cmd/analysis_worker` | Worker de análise das entrevistas |
| `internal/domain` | Entidades, interfaces e regras de negócio |
| `internal/services` e `usecase` | Casos de uso e orquestração do diálogo |
| `internal/infrastructure` | Integrações externas (Deepgram, OpenAI, ElevenLabs, LiveKit) |
| `internal/repository` | Persistência em PostgreSQL com `pgx` |

- **Multi-tenant:** um schema PostgreSQL por cliente; o pool configura o `search_path` a partir do tenant do JWT.
- **Observabilidade:** logs estruturados em JSON com `zap` e TraceID por requisição.
- **Segurança:** middlewares de security headers, recoverer e request logger; pool de conexões com timeouts.
- **API documentada** com Swagger (`docs/`).

## Stack

Go · Chi · PostgreSQL (`pgx`) · Redis · LiveKit · Deepgram · OpenAI · ElevenLabs · Templ · Alpine.js · Docker

## Executando localmente

Pré-requisitos: Docker, Docker Compose e Go 1.21+.

```bash
cp .env.example .env     # preencha com as suas chaves; nunca versione o .env
docker-compose up -d     # PostgreSQL, Redis e LiveKit
```

Os passos de cada serviço (server, worker e analysis_worker) e de deploy estão em [`ARCHITECTURE.md`](ARCHITECTURE.md), [`DEPLOY.md`](DEPLOY.md) e [`DEPLOY_RAILWAY.md`](DEPLOY_RAILWAY.md).

## Autor

**Wallace Leonne Cardoso de Souza** · [APOC Tecnologia](https://apoctecnologia.com.br) · [LinkedIn](https://www.linkedin.com/in/wallace-leonne-cardoso-de-souza/)
