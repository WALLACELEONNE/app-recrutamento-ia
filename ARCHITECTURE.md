# Documentação da Arquitetura e APIs

## 1. Visão Geral da Arquitetura
A plataforma é um SaaS (multi-tenant) desenhado para realizar entrevistas automatizadas utilizando Inteligência Artificial por voz em tempo real. Toda a aplicação foi construída em **Go (Golang)** garantindo performance, simultaneidade via goroutines e baixo custo de infraestrutura.

A arquitetura adota princípios de **Clean Architecture** e separa as responsabilidades da seguinte forma:
- `cmd/server`: Servidor HTTP e painel web (RH e setup do candidato).
- `cmd/worker`: Worker de IA que se conecta às salas WebRTC (LiveKit) e orquestra STT, LLM e TTS.
- `internal/domain`: Entidades core, interfaces e regras de negócio.
- `internal/infrastructure`: Integrações com APIs externas (Deepgram, OpenAI, ElevenLabs) e LiveKit.
- `internal/services`: Casos de uso e orquestração (ex: fluxo de fala do candidato vs IA).
- `internal/repository`: Camada de persistência (PostgreSQL usando `pgx`).

---

## 2. Orquestração em Tempo Real (AI Worker)

O AI Worker é o coração da plataforma. Ele usa o `LiveKit Server SDK` para entrar na mesma "sala" virtual que o candidato.

**Fluxo de Dados (Full-Duplex):**
1. O candidato fala no microfone (Browser). O LiveKit roteia os pacotes RTP para o Worker.
2. O Worker extrai o Payload (PCM/Opus) e envia para o **Deepgram** (STT) via WebSocket (`TranscribeStream`).
3. Ao receber o texto final (IsFinal=true), o Worker envia a string e o histórico da conversa para a **OpenAI (GPT-4o)** (`GenerateResponseStream`).
4. Os tokens de texto gerados pela OpenAI são enviados incrementalmente (stream) para a **ElevenLabs** (TTS).
5. O áudio gerado pela ElevenLabs é transformado em samples (`media.Sample`) e escrito de volta na trilha de áudio da IA (`aiTrack.WriteSample`), tocando instantaneamente no fone do candidato.

**Meta de Latência**: O uso de streaming de ponta a ponta (Deepgram -> OpenAI -> ElevenLabs -> LiveKit) permite manter a latência abaixo de 1.2 segundos.

---

## 3. Modelagem Multi-Tenant e Segurança

- **Isolamento de Dados**: Utiliza-se a abordagem de múltiplos schemas no PostgreSQL (`schema_acme`, `schema_techcorp`), onde o pool de conexão configura dinamicamente o `search_path` antes das transações baseadas no Tenant ID extraído do JWT.
- **Tratamento de Erros e Logs**: Utiliza o `go.uber.org/zap` para logs estruturados em JSON, injetando Caller, TraceID (do Chi Middleware) para facilitar observabilidade.
- **Validação e Segurança**: 
  - Middlewares implementados: `SecurityHeaders`, `RequestLogger`, `Recoverer`.
  - Conexão DB (`pgxpool`): Configurada com timeouts, lifetime de conexões (1h) e `ping` inicial.

---

## 4. Endpoints da API (Server) - Estrutura Base

| Método | Rota | Descrição | Auth |
|---|---|---|---|
| `GET` | `/health` | Liveness probe do Kubernetes / Docker | Não |
| `GET` | `/api/v1/status` | Verifica status de conectividade (DB, Redis) | Não |
| `POST` | `/api/v1/interviews/link` | RH gera link único de entrevista | JWT (RH) |
| `GET` | `/api/v1/interviews/{token}`| Retorna metadados para setup do candidato (Microfone) | Token |

*Nota: A interface web será acoplada diretamente ao Server utilizando Go Templ (SSR) e Alpine.js.*

---

## 5. Como Executar Localmente

### Pré-requisitos
- Docker & Docker Compose
- Go 1.21+

### Serviços Locais (Docker)
Levante o banco de dados, Redis e LiveKit Server rodando:
```bash
docker-compose up -d
```

### Inicializar os Microserviços
**Terminal 1 (Server):**
```bash
cp .env.example .env # (configure suas chaves)
go run ./cmd/server/main.go
```

**Terminal 2 (Worker de IA):**
```bash
go run ./cmd/worker/main.go
```

---
*Criado seguindo os princípios de Software Engineering e Padrões Solid aplicados em Go.*
