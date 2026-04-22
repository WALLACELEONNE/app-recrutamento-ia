# Guia de Execução Local e Deploy (Docker & Railway)

Este documento descreve os comandos detalhados para rodar a aplicação em um ambiente local isolado (usando Docker e Docker Compose) e como fazer o deploy na plataforma [Railway](https://railway.app/).

## 1. Execução Local com Docker Compose

A infraestrutura local orquestra **6 serviços**:
- `postgres` (Banco de Dados)
- `redis` (Cache / Sessões)
- `livekit` (Servidor WebRTC/WebSocket)
- `server` (A aplicação principal HTTP em Golang)
- `worker` (O processo de IA que consome LiveKit)
- `nginx` (Proxy Reverso e Load Balancer)

### Escalabilidade Horizontal Local

O arquivo `docker-compose.yml` está preparado para escalar instâncias de `server` e `worker` sob demanda, pois removemos portas hardcoded em favor do Nginx roteando o tráfego interno do Docker.

1. **Inicie os containers e escale o servidor para 3 instâncias**:
   ```bash
   docker-compose up --build --scale server=3 -d
   ```
   *O Nginx fará o round-robin automático das requisições na porta 3000 para as três réplicas do Server.*

2. **Acompanhar os logs centralizados**:
   O Docker está configurado com limites de tamanho de log (`json-file`, max 10m) para evitar que seu disco fique cheio.
   Para ver todos os logs em tempo real:
   ```bash
   docker-compose logs -f
   ```
   Para ver o log de um serviço específico (ex: worker):
   ```bash
   docker-compose logs -f worker
   ```

4. **Verificar a saúde dos serviços**:
   O `docker-compose.yml` está configurado com Health Checks rigorosos.
   ```bash
   docker ps
   ```
   *(Você deverá ver `(healthy)` na coluna STATUS para postgres, redis, livekit, server e worker)*

5. **Parar e limpar o ambiente**:
   Para parar e remover os containers e a rede (mantendo os dados persistidos nos volumes):
   ```bash
   docker-compose down
   ```
   Para deletar também os volumes de dados (Resetar Banco de Dados):
   ```bash
   docker-compose down -v
   ```

---

## 2. Deploy no Railway

O projeto utiliza um pipeline automatizado de CI/CD com o GitHub Actions, mas você também pode gerenciar o deploy e os logs utilizando a Railway CLI.

### Pré-requisitos
- Ter o Node.js e o npm instalados.
- Instalar a CLI do Railway:
  ```bash
  npm i -g @railway/cli
  ```

### Autenticação e Configuração

1. **Fazer Login no Railway**:
   ```bash
   railway login
   ```
   *(Isso abrirá uma janela no navegador para você autorizar o dispositivo)*

2. **Linkar seu repositório local a um projeto do Railway**:
   ```bash
   railway link
   ```
   *(Siga as instruções no terminal para selecionar seu projeto)*

### Gerenciamento de Variáveis de Ambiente

As variáveis de produção nunca devem ir para o GitHub. Para configurá-las no Railway:
1. Abra a interface web do projeto ou use o terminal:
   ```bash
   # Visualizar as variáveis atuais do ambiente
   railway vars
   
   # Configurar uma nova variável
   railway vars set APP_ENV=production PORT=3000
   ```

### Deploy Manual via CLI

Caso queira fazer um deploy ignorando o GitHub Actions (por exemplo, testando código local não commitado):
```bash
railway up
```
Isso utilizará o arquivo `railway.json` que aponta para o `Dockerfile.server` por padrão.

### Estrutura Multi-Serviço no Railway

O arquivo `railway.json` principal está configurado para fazer deploy do **Server**. 
Para o **Worker** (que utiliza o `Dockerfile.worker`), você precisa criar um **Service** secundário dentro do seu projeto no painel do Railway e configurá-lo para usar a `Custom Dockerfile` apontando para `/Dockerfile.worker` sem expor portas HTTP externas.

### Visualizando Logs de Produção

Para ver os logs do último deploy em tempo real no terminal:
```bash
railway logs
```

---

## 3. Pipeline de CI/CD (GitHub Actions)

A cada `push` na branch `main`, a Action `.github/workflows/deploy.yml` é disparada:
1. **Testes**: Executa o linter e todos os testes unitários (`go test -v -coverprofile=coverage.out ./...`).
2. **Cobertura**: Valida se a cobertura de código é superior a **80%**. Se não for, o deploy é cancelado.
3. **Deploy Automatizado**: Se tudo passar, a action usa o comando `railway up --detach` injetando a `RAILWAY_TOKEN` definida no seu GitHub Secrets.

**Para habilitar o CI/CD:**
1. No seu dashboard do Railway, crie um Token de API.
2. Vá nas configurações do repositório do GitHub -> Secrets and variables -> Actions.
3. Adicione uma nova secret com o nome `RAILWAY_TOKEN` e cole o token gerado.
