# PLANO DE IMPLEMENTAÇÃO: Plataforma de Recrutamento com IA por Voz

## Stack Alternativa: Go + LiveKit + Templ

---

## 1. OBJETIVOS

### 1.1 Objetivo Principal
Implementar uma plataforma SaaS de entrevistas por voz conduzidas por IA, utilizando Go como linguagem única, LiveKit para infraestrutura de áudio/vídeo em tempo real, e Templ para renderização server-side.

### 1.2 Objetivos Específicos

| # | Objetivo | Métrica de Sucesso |
|---|----------|-------------------|
| 1 | Migrar/criar frontend com Go + Templ SSR | Pages carregam em <200ms |
| 2 | Integrar LiveKit para áudio/voz | Entrevistas funcionam em tempo real |
| 3 | Implementar Worker Go como IA | Latência < 1.2s ponta a ponta |
| 4 | Manter isolamento multi-tenant por schema PostgreSQL | Cada empresa acessa apenas seus dados |
| 5 | Criar fluxo completo de entrevista (setup → turno → análise) | End-to-end funcional |
| 6 | Implementar Analysis Engine para scoring | Relatórios gerados automaticamente |
| 7 | Garantir conformidade LGPD | Dados de candidatos protegidos |

---

## 2. ARQUITETURA ALVO

```
┌─────────────────────────────────────────────────────────────────────┐
│                          CLIENTES                                    │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────────┐ │
│  │   Empresa   │    │  Candidato  │    │       Admin SaaS        │ │
│  │  (Dashboard)│    │ (Entrevista) │    │      (Backoffice)       │ │
│  │  Templ/Go   │    │  Templ/Go   │    │       Templ/Go          │ │
│  │  Alpine.js  │    │  Alpine.js  │    │       Alpine.js         │ │
│  └──────┬──────┘    └──────┬──────┘    └───────────┬─────────────┘ │
└─────────┼──────────────────┼───────────────────────┼───────────────┘
          │                  │                       │
          ▼                  ▼                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                    LIVEKIT CLOUD / SELF-HOSTED                       │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                    LiveKit Server (Go)                           ││
│  │   Room: interview-{session_id}   Participants: [candidato, ia]   ││
│  └─────────────────────────────────────────────────────────────────┘│
│                              │                                       │
│                              ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │                    AI WORKER (Go)                                ││
│  │  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐  ││
│  │  │   VAD    │───▶│   STT    │───▶│   LLM    │───▶│   TTS    │  ││
│  │  │(silence) │    │(Deepgram)│    │(GPT-4o)  │    │(ElevenLabs│  ││
│  │  └──────────┘    └──────────┘    └──────────┘    └──────────┘  ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         DADOS & STORAGE                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐               │
│  │ PostgreSQL   │  │    Redis     │  │     S3       │               │
│  │(multi-tenant)│  │  (sessions)  │  │  (áudio/pdfs)│               │
│  └──────────────┘  └──────────────┘  └──────────────┘               │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. FASES DE IMPLEMENTAÇÃO

### FASE 1: Infraestrutura Base (Semana 1)
**Duração estimada**: 5 dias
**Responsável**: Backend/DevOps

#### 3.1.1 Configuração do Projeto Go

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 1.1 | Inicializar módulo Go com `go mod init` | `go.mod` com dependências |
| 1.2 | Configurar estrutura de diretórios (cmd, internal, pkg) | Estrutura de pastas padronizada |
| 1.3 | Configurar `templ` e templates base | Templates HTML compilando |
| 1.4 | Configurar Tailwind CSS | Estilização funcionando |
| 1.5 | Setup inicial do LiveKit Go SDK | Cliente LiveKit configurado |

**Estrutura de diretórios:**
```
app-recrutamento-ia/
├── cmd/
│   ├── server/           # Servidor HTTP principal
│   └── worker/           # AI Worker (LiveKit listener)
├── internal/
│   ├── domain/           # Entidades e interfaces
│   ├── handlers/         # Handlers HTTP (Templ)
│   ├── infrastructure/
│   │   ├── livekit/      # Cliente LiveKit
│   │   ├── openai/       # Cliente OpenAI
│   │   ├── deepgram/     # Cliente Deepgram
│   │   └── elevenlabs/   # Cliente ElevenLabs
│   ├── repository/       # Acesso a dados PostgreSQL
│   └── services/         # Lógica de negócio
├── migrations/           # Migrations SQL
├── templates/            # Arquivos .templ
├── static/              # CSS, JS, imagens
└── docker-compose.yml   # Desenvolvimento local
```

#### 3.1.2 Configuração do LiveKit

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 1.6 | Criar conta LiveKit Cloud (ou self-hosted) | Credenciais obtidas |
| 1.7 | Configurar `livekit.yaml` no projeto | Arquivo de configuração |
| 1.8 | Implementar conexão básica ao LiveKit | Room criado com sucesso |
| 1.9 | Testar publish/subscribe de áudio | Áudio fluindo entre participantes |

#### 3.1.3 PostgreSQL Multi-Tenant

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 1.10 | Setup PostgreSQL via Docker | PostgreSQL rodando localmente |
| 1.11 | Criar migrations do schema público | Tabelas `organizations`, `users` |
| 1.12 | Implementar função de criação de schema por tenant | `CREATE SCHEMA schema_{slug}` |
| 1.13 | Implementar middleware de search_path | Isolamento automático por tenant |

---

### FASE 2: Frontend com Templ (Semana 2)
**Duração estimada**: 5 dias
**Responsável**: Frontend/Full-stack

#### 3.2.1 Templates Base

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 2.1 | Criar layout base (header, footer, container) | `layout.templ` |
| 2.2 | Implementar página de login/auth | `login.templ` |
| 2.3 | Criar dashboard RH (lista de vagas) | `dashboard.templ` |
| 2.4 | Implementar página de detalhes da vaga | `job_detail.templ` |
| 2.5 | Criar modal de convite de candidato | `invite_modal.templ` (Alpine.js) |

#### 3.2.2 Página de Entrevista (Player)

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 2.6 | Criar página de entrevista (`/interview/{session_id}`) | Template de entrevista |
| 2.7 | Integrar Alpine.js para controles de áudio | Play/pause/mute buttons |
| 2.8 | Implementar visualização de estado (listening/speaking) | Indicadores visuais |
| 2.9 | Adicionar indicador de tempo de entrevista | Cronômetro funcional |
| 2.10 | Criar tela de encerramento e redirecionamento | `interview_end.templ` |

#### 3.2.3 Dashboard Admin

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 2.11 | Criar área admin (`/admin`) | Painel administrativo |
| 2.12 | Listar organizações/tenants | `admin/organizations.templ` |
| 2.13 | Visualizar métricas gerais | Dashboard com métricas |

---

### FASE 3: Worker de IA (Semana 3)
**Duração estimada**: 5 dias
**Responsável**: Backend/IA

#### 3.3.1 Pipeline de Transcrição (STT)

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 3.1 | Configurar cliente Deepgram (STT streaming) | Cliente Deepgram em Go |
| 3.2 | Implementar subscribe no track de áudio do candidato | Stream de áudio chegando |
| 3.3 | Implementar VAD (Voice Activity Detection) | Detecção de silêncio funciona |
| 3.4 | Integrar com LiveKit (audio track → Deepgram) | Transcrição em tempo real |

**Diagrama do fluxo:**
```
LiveKit Room
    │
    ▼
[Audio Track - Candidato]
    │
    ▼ (subscribe)
[Go Worker]
    │
    ▼
[Deepgram Streaming API]
    │
    ▼ (texto transcrito)
[Processamento de turno]
```

#### 3.3.2 Integração com LLM (GPT-4o)

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 3.5 | Configurar cliente OpenAI (streaming responses) | Cliente OpenAI configurado |
| 3.6 | Implementar construção de prompt (persona + contexto) | Prompt dinâmico por vaga |
| 3.7 | Implementar histórico de conversa | Contexto mantido entre turnos |
| 3.8 | Implementar streaming de resposta | Tokens chegam em tempo real |

**Estrutura do Prompt:**
```go
type InterviewPrompt struct {
    Persona      string            // "Entrevistadora experiente de RH"
    JobContext   string            // Descrição da vaga
    History      []Turn           // Histórico de turnos
    CurrentTurn  string            // Última fala do candidato
}

type Turn struct {
    Role     string    // "ai" ou "candidate"
    Content  string    // Texto da fala
    Timestamp int64    // Quando ocorreu
}
```

#### 3.3.3 Síntese de Voz (TTS)

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 3.9 | Configurar cliente ElevenLabs | Cliente ElevenLabs configurado |
| 3.10 | Implementar TTS streaming (chunk por sentença) | Áudio gerado incrementalmente |
| 3.11 | Publicar audio track no LiveKit (voz da IA) | Candidato ouve a resposta |
| 3.12 | Implementar controle de turno (não falar enquanto candidato fala) | Máquina de estados |

**Estados do Turno:**
```
IDLE ──(candidato começa a falar)──▶ LISTENING
LISTENING ──(silêncio detectado)──▶ PROCESSING
PROCESSING ──(LLM responde)──▶ SPEAKING
SPEAKING ──(fim da resposta)──▶ IDLE
```

#### 3.3.4 Worker Completo

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 3.13 | Integrar STT + LLM + TTS em um worker coeso | Worker funcional completo |
| 3.14 | Implementar graceful shutdown | Worker para corretamente |
| 3.15 | Adicionar logs e tracing | Observabilidade básica |

---

### FASE 4: Backend e Lógica de Negócio (Semana 4)
**Duração estimada**: 5 dias
**Responsável**: Backend

#### 3.4.1 API REST / HTTP Handlers

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 4.1 | Implementar CRUD de Organizations | `POST/GET/PUT organizations` |
| 4.2 | Implementar CRUD de Jobs (vagas) | `POST/GET/PUT jobs` |
| 4.3 | Implementar CRUD de Candidates | `POST/GET candidates` |
| 4.4 | Implementar criação de Sessions | `POST interview_sessions` |
| 4.5 | Implementar convite por email | Token JWT + link único |

#### 3.4.2 Autenticação e Autorização

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 4.6 | Implementar JWT auth | Token criado e validado |
| 4.7 | Implementar middleware de tenant | `org_id` extraído do JWT |
| 4.8 | Implementar roles (admin, recruiter, viewer) | Permissões por role |
| 4.9 | Proteger rotas com auth middleware | Rotas autenticadas |

#### 3.4.3 Integração Worker ↔ Backend

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 4.10 | Worker busca contexto da vaga ao iniciar sessão | Job info carregada |
| 4.11 | Worker salva transcrição no PostgreSQL | `session_turns` preenchido |
| 4.12 | Worker atualiza status da sessão | `interview_sessions.status` |

---

### FASE 5: Analysis Engine (Semana 5)
**Duração estimada**: 5 dias
**Responsável**: Backend/IA

#### 3.5.1 Processamento Pós-Sessão

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 5.1 | Detectar fim de sessão (worker recebe evento) | Fim detectado automaticamente |
| 5.2 | Buscar transcrição completa da sessão | Todo o texto recuperado |
| 5.3 | Enviar para análise via LLM (GPT-4o) | Análise processada |

#### 3.5.2 Scoring e Relatório

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 5.4 | Implementar prompt de scoring | Avaliação por competência |
| 5.5 | Calcular `overall_score` (0-100) | Score numérico |
| 5.6 | Extrair competências (JSONB) | `competencies` por skill |
| 5.7 | Gerar `summary_text` | Resumo em texto |
| 5.8 | Gerar relatório PDF | `report_s3_key` |

#### 3.5.3 Webhook para Empresa

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 5.9 | Implementar webhook notification | Notificação HTTP POST |
| 5.10 | Enviar relatório para empresa | Dados transmitidos |

---

### FASE 6: Testes e Validação (Semana 6)
**Duração estimada**: 5 dias
**Responsável**: QA/Todos

#### 3.6.1 Testes Unitários

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 6.1 | Testar handlers HTTP | Handlers testados |
| 6.2 | Testar repositories | Queries testadas |
| 6.3 | Testar serviços de domínio | Lógica de negócio testada |
| 6.4 | Testar worker (STT → LLM → TTS isolado) | Pipeline IA testado |

#### 3.6.2 Testes de Integração

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 6.5 | Testar fluxo completo de entrevista | End-to-end testado |
| 6.6 | Testar isolamento multi-tenant | Tenant A não vê dados de B |
| 6.7 | Testar LiveKit em ambiente local | WebRTC funciona |

#### 3.6.3 Testes de Carga e Performance

| Tarefa | Descrição | Entregável |
|--------|-----------|------------|
| 6.8 | Medir latência STT (Deepgram) | Latência < 500ms |
| 6.9 | Medir latência LLM (GPT-4o streaming) | First token < 500ms |
| 6.10 | Medir latência TTS (ElevenLabs) | Latência < 300ms |
| 6.11 | Medir latência total (turno completo) | Total < 1.2s |

---

## 4. CRITÉRIOS DE SUCESSO

### 4.1 Critérios Técnicos

| Critério | Meta | Como Medir |
|----------|------|------------|
| Latência STT | < 500ms | Logs do worker |
| Latência LLM (first token) | < 500ms | Métricas OpenAI |
| Latência TTS | < 300ms | Logs ElevenLabs |
| Latência total ponta a ponta | < 1.2s | Teste manual com cronômetro |
| Uptime do sistema | > 99% | Monitoramento |
| Cobertura de testes | > 70% | `go test -cover` |

### 4.2 Critérios Funcionais

| Critério | Descrição | Validação |
|----------|-----------|-----------|
| F1 | Candidato consegue fazer entrevista por voz completa | Teste manual |
| F2 | IA faz perguntas baseadas na vaga | Verificação de prompts |
| F3 | Relatório é gerado após entrevista | Verificação no dashboard |
| F4 | Empresa consegue ver relatório no dashboard | UI funciona |
| F5 | Isolamento de dados entre tenants | Teste de segurança |
| F6 | Conformidade LGPD (anonimização) | Dados protegidos |

### 4.3 Critérios de Qualidade

| Critério | Descrição |
|----------|-----------|
| Q1 | Código compila sem erros (`go build ./...`) |
| Q2 | Linting passa (`golangci-lint run`) |
| Q3 | Tests passam (`go test ./...`) |
| Q4 | Templates compilados sem warnings (`templ generate`) |
| Q5 | Sem secrets no código (credenciais em env vars) |

---

## 5. RESTRIÇÕES TÉCNICAS

### 5.1 Restrições de Infraestrutura

| Restrição | Descrição |
|-----------|-----------|
| R1 | LiveKit: Necessário API key (Cloud ou self-hosted) |
| R2 | OpenAI: Necessário API key com quota disponível |
| R3 | Deepgram: Necessário API key para STT streaming |
| R4 | ElevenLabs: Necessário API key para TTS |
| R5 | PostgreSQL: Mínimo 4GB RAM para multi-tenant |
| R6 | Redis: Necessário para sessão do LiveKit (se usar) |

### 5.2 Restrições de Desenvolvimento

| Restrição | Descrição |
|-----------|-----------|
| R7 | Go 1.21+ (para usar generics e melhores performance) |
| R8 | Templ v0.2+ (versão mais recente) |
| R9 | LiveKit Go SDK v1.0+ |
| R10 | Node.js NÃO pode ser usado (stack Go pura) |

### 5.3 Restrições de Prazo

| Fase | Prazo | Entrega |
|------|-------|---------|
| Fases 1-2 | Semana 2 | MVP de frontend e LiveKit básico |
| Fases 3-4 | Semana 4 | Worker de IA + backend funcional |
| Fases 5-6 | Semana 6 | Sistema completo testado |

---

## 6. COMO VALIDAR RESULTADOS

### 6.1 Validação por Fase

**Fase 1 (Infraestrutura):**
- [ ] `go build ./cmd/server` compila
- [ ] `docker-compose up` sobe PostgreSQL e Redis
- [ ] LiveKit client conecta ao servidor
- [ ] Migrations criam schema público

**Fase 2 (Frontend):**
- [ ] `templ generate` compila templates
- [ ] `go run ./cmd/server` inicia servidor
- [ ] Página de login carrega (`/login`)
- [ ] Dashboard carrega (`/dashboard`)
- [ ] Alpine.js modais abrem/fecham

**Fase 3 (Worker IA):**
- [ ] Worker conecta ao LiveKit room
- [ ] STT transcreve áudio em tempo real
- [ ] LLM responde com streaming
- [ ] TTS gera áudio e pubblica no LiveKit
- [ ] Máquina de estados controla turno

**Fase 4 (Backend):**
- [ ] CRUD de organizations funciona
- [ ] CRUD de jobs funciona
- [ ] Convite de candidato enviado
- [ ] JWT auth protege rotas
- [ ] Multi-tenant isolado (schema por org)

**Fase 5 (Analysis):**
- [ ] Fim de sessão detectado
- [ ] Análise processada (GPT-4o)
- [ ] Score calculado (0-100)
- [ ] Relatório PDF gerado
- [ ] Webhook enviado

**Fase 6 (Testes):**
- [ ] `go test ./...` passa
- [ ] Cobertura > 70%
- [ ] Latência < 1.2s medida
- [ ] Teste manual de entrevista completo

### 6.2 Checklist de Aceitação

```
□ Servidor HTTP inicia na porta 3000
□ Templates renderizam HTML corretamente
□ LiveKit room é criado para cada sessão
□ Áudio flui: Candidato → Worker → IA → Candidato
□ Transcrição é salva no PostgreSQL
□ Análise é gerada após entrevista
□ Empresa vê relatório no dashboard
□ Tenant A NÃO vê dados do Tenant B
```

---

## 7. ENTREGÁVEIS

### 7.1 Entregáveis de Código

| # | Entregável | Local | Descrição |
|---|------------|-------|-----------|
| E1 | Servidor HTTP | `cmd/server/main.go` | App principal com Templ |
| E2 | AI Worker | `cmd/worker/main.go` | Worker LiveKit → OpenAI |
| E3 | Templates | `templates/*.templ` | Páginas HTML |
| E4 | Handlers | `internal/handlers/` | HTTP handlers |
| E5 | Repositories | `internal/repository/` | Acesso a dados |
| E6 | LiveKit Client | `internal/infrastructure/livekit/` | Integração LiveKit |
| E7 | AI Clients | `internal/infrastructure/` | OpenAI, Deepgram, ElevenLabs |
| E8 | Migrations | `migrations/` | SQL schema |

### 7.2 Entregáveis de Documentação

| # | Entregável | Descrição |
|---|------------|-----------|
| D1 | README.md | Setup e instruções de uso |
| D2 | .env.example | Variáveis de ambiente |
| D3 | docker-compose.yml | Ambiente de desenvolvimento |
| D4 | API Documentation | Endpoints e payloads |
| D5 | ARCHITECTURE.md | Decisões de arquitetura |

### 7.3 Entregáveis de Testes

| # | Entregável | Descrição |
|---|------------|-----------|
| T1 | Testes unitários | `*_test.go` por pacote |
| T2 | Testes de integração | `integration/` |
| T3 | Teste E2E manual | Script de teste |

---

## 8. PRÓXIMOS PASSOS

### Imediato (Hoje)
1. [ ] Confirmar tecnologias (LiveKit Cloud vs Self-hosted)
2. [ ] Obter API keys (OpenAI, Deepgram, ElevenLabs, LiveKit)
3. [ ] Inicializar repositório Git
4. [ ] Criar estrutura de diretórios

### Esta Semana (Fase 1)
1. [ ] Configurar `go.mod` com todas dependências
2. [ ] Setup `docker-compose.yml` (PostgreSQL + Redis)
3. [ ] Implementar templates base com Templ
4. [ ] Configurar Tailwind CSS
5. [ ] Testar LiveKit SDK

---

## 9. CONTATO E SUPORTE

Para dúvidas durante a implementação, consulte:
- **Documentação LiveKit**: https://docs.livekit.io
- **Documentação Templ**: https://templ.guide
- **Documentação Go**: https://go.dev/doc

---

## 10. MODELO DE DADOS (RESUMO)

### Schema Público (compartilhado)
```
organizations
├── id (uuid PK)
├── slug (varchar) - ex: acme-corp
├── plan (enum: starter/growth/enterprise)
├── schema_name (varchar)
└── created_at

users (global)
├── id (uuid PK)
├── email
├── password_hash
├── org_id (uuid FK → organizations)
└── role (enum: admin/recruiter/viewer)
```

### Schema por Tenant (ex: schema_acme)
```
jobs
├── id (uuid PK)
├── title
├── department
├── interview_config (jsonb)
│   ├── n_questions
│   ├── max_minutes
│   ├── persona
│   ├── tone
│   └── focus_areas[]
└── ...

interview_templates
├── id (uuid PK)
├── job_id (uuid FK → jobs)
├── system_prompt (text)
├── opening_message (text)
├── question_bank (jsonb[])
├── version (int)
└── is_active (bool)

candidates
├── id (uuid PK)
├── name, email, phone
├── job_id (uuid FK → jobs)
├── invite_token (uuid)
├── expires_at
├── gdpr_consent_at
└── anonymized_at

interview_sessions
├── id (uuid PK)
├── candidate_id (FK)
├── job_id (FK)
├── status (enum: invited/in_progress/done)
├── started_at, ended_at, duration_s
├── audio_s3_key
└── transcript_s3_key

session_turns
├── id (uuid PK)
├── session_id (FK)
├── role (enum: ai/candidate)
├── content (text)
├── turn_index (int)
├── audio_offset_ms
└── duration_ms

analysis_results
├── id (uuid PK)
├── session_id (FK, unique)
├── overall_score (numeric 4,2)
├── competencies (jsonb)
├── summary_text (text)
└── report_s3_key
```

---

## 11. STACK TÉCNICA (DEFINITIVA)

| Camada | Tecnologia | Versão Mínima |
|--------|------------|---------------|
| Linguagem | Go | 1.21+ |
| Templates | Templ | 0.2+ |
| CSS | Tailwind CSS | 3.0+ |
| Interatividade | Alpine.js | 3.0+ |
| WebRTC/Áudio | LiveKit | Server + Go SDK 1.0+ |
| STT | Deepgram | Streaming API |
| LLM | OpenAI | GPT-4o |
| TTS | ElevenLabs | Streaming API |
| Database | PostgreSQL | 15+ |
| Cache/Sessions | Redis | 7+ |
| Object Storage | S3 (ou compatível) | - |
| Auth | JWT | - |
| Deployment | Docker + Compose | - |

---

*Documento criado em: 2026-04-21*
*Plano de Implementação - Plataforma de Recrutamento com IA por Voz*
