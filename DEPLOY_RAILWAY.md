# Documentação de Deploy (Railway & Local Docker)

Este documento descreve como a infraestrutura do projeto está configurada para deploy na plataforma **Railway**, garantindo consistência com o ambiente local via **Docker**.

## 1. Topologia de Serviços no Railway

O projeto utiliza um repositório Monorepo (Golang) com múltiplos pontos de entrada. Para isso, no Railway, provisionamos **5 serviços** no mesmo projeto:

- **Postgres**: Banco de dados relacional.
- **Redis**: Serviço de mensageria e cache.
- **server**: O backend e frontend (HTTP).
- **worker**: Worker para WebRTC e LiveKit.
- **analysis_worker**: Worker para análise assíncrona com LLM e filas.

> **Nota:** Todos os serviços apontam para o mesmo repositório GitHub, mas o CI/CD faz o trigger de qual serviço executar baseado no `Dockerfile` e comando (`railway up --service <nome>`).

## 2. Configurações de Timezone e Encoding (UTF-8)

Para evitar problemas de renderização de caracteres acentuados, símbolos especiais e horários incorretos de banco de dados, estabelecemos como **padrão universal**:
- **Timezone**: `GMT-3` (America/Sao_Paulo)
- **Encoding**: `UTF-8`

### Como isso foi implementado?

#### a) Dockerfiles
Os Dockerfiles de todos os serviços baseados em Alpine Linux (`Dockerfile.server`, `Dockerfile.worker`, `Dockerfile.analysis_worker`) possuem o pacote `tzdata` instalado.

```dockerfile
# Adiciona certificados SSL e fuso horário
RUN apk --no-cache add ca-certificates tzdata

# Configuração de Timezone (Horário de Brasília) e Encoding (UTF-8)
ENV TZ=America/Sao_Paulo
ENV LANG=C.UTF-8
ENV LC_ALL=C.UTF-8
```

#### b) Local (docker-compose.yml)
Todos os serviços executados localmente (`postgres`, `redis`, `nginx`) receberam injetadas as mesmas variáveis de ambiente:
```yaml
environment:
  - TZ=America/Sao_Paulo
  - LANG=C.UTF-8
  - LC_ALL=C.UTF-8
```

#### c) Railway (Variáveis de Ambiente)
Via CLI / Painel do Railway, configuramos em **TODOS** os 5 serviços as seguintes *Shared Variables*:
- `TZ` = `America/Sao_Paulo`
- `LANG` = `C.UTF-8`
- `LC_ALL` = `C.UTF-8`

## 3. Comunicação entre Serviços no Railway

O Railway permite injeção automática de variáveis de um serviço em outro.

Nos serviços em Golang (`server`, `worker`, `analysis_worker`), as credenciais de banco e redis foram configuradas via Railway Reference Variables:
- `DATABASE_URL` = `${{ Postgres.DATABASE_URL }}`
- `REDIS_URL` = `${{ Redis.REDIS_URL }}`

## 4. Pipeline CI/CD (GitHub Actions)

Toda vez que a branch `main` recebe um *Push* (ou PR aprovado), a action definida em `.github/workflows/deploy.yml`:
1. Roda os testes unitários em Go.
2. Faz validação de Linting e Cobertura.
3. Executa a CLI do Railway em background e faz o deploy direcionado:
   ```bash
   railway up --service server --detach
   railway up --service worker --detach
   railway up --service analysis_worker --detach
   ```

A action é autorizada por um único secret no repositório: `RAILWAY_TOKEN`.

## 5. Manutenção e Troubleshooting

- **Logs do Railway**: Se os logs estiverem com caracteres estranhos, certifique-se de que a variável `LANG=C.UTF-8` está corretamente atribuída nas configurações (Variables) do serviço afetado.
- **Timestamp de banco**: Ao persistir datas via GORM/PGX, o Postgres utilizará nativamente o fuso `America/Sao_Paulo` devido à injeção da variável `TZ`.
- **Rotas Públicas**: O serviço `server` expõe uma URL pública. Já os workers e bancos não precisam de IPs públicos, operando internamente pela rede privada da infraestrutura do Railway (`*.railway.internal`).
