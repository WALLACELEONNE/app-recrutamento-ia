---
name: "ai-interview-pipeline"
description: "Diagnoses and configures the AI Interview audio pipeline (LiveKit, WebRTC, Deepgram STT, OpenAI LLM/TTS). Invoke when troubleshooting voice connections, silence, or audio codec errors."
---

# AI Interview Pipeline Troubleshooter

Esta skill deve ser ativada sempre que o usuário relatar problemas na sala de entrevista de IA, como falta de áudio, conexão caindo, IA não respondendo, ou erros no console relacionados ao WebRTC e Alpine.js.

## Arquitetura do Pipeline

A entrevista por voz segue este caminho exato:
1. **Frontend (Alpine.js)** captura o áudio do microfone do candidato.
2. **LiveKit Cloud** atua como SFU (WebRTC) transmitindo os pacotes.
3. **Go Worker (Backend)** se conecta à mesma sala do LiveKit usando `server-sdk-go`.
4. **Deepgram (STT)** recebe os pacotes Opus brutos via WebSocket e transcreve a fala do candidato (`pion/webrtc` track -> `STTClient`).
5. **OpenAI (LLM)** recebe o texto transcrito, gera uma resposta contextualizada baseada na vaga e histórico (`LLMClient`).
6. **OpenAI (TTS)** transforma o texto de resposta em um áudio codificado em Opus (OGG).
7. **OggReader** decodifica os pacotes OGG e escreve na trilha de áudio local do Worker (`LocalSampleTrack`), que é então enviada de volta pelo LiveKit para o candidato ouvir.

## Diagnóstico Passo a Passo (Troubleshooting Guide)

Quando o usuário relatar falhas, siga este checklist rigorosamente:

### 1. Problemas de Frontend (DataCloneError / Alpine.js)
- **Sintoma:** O console do navegador exibe `DataCloneError: Failed to execute 'structuredClone'` ao mutar/desmutar o microfone.
- **Causa:** O Alpine.js tenta transformar objetos complexos do LiveKit (`LivekitClient.Room`, `audioElement`) em proxies reativos, quebrando-os.
- **Solução:** Variáveis do LiveKit devem ser declaradas **fora** do `return { ... }` do componente Alpine.js (como variáveis locais de closure).

### 2. Problemas de STT (Deepgram não reconhece o áudio)
- **Sintoma:** O Worker crasha ou o log exibe `Deepgram error {"msg": "close 1011"}`.
- **Causa:** O LiveKit envia os pacotes no formato Opus bruto, mas o Deepgram espera Linear16 (PCM) por padrão, ou os parâmetros não foram especificados.
- **Solução:** A conexão WebSocket com o Deepgram deve especificar:
  ```go
  Encoding: "opus",
  SampleRate: 48000,
  Channels: 1,
  ```

### 3. Problemas de TTS (IA não fala ou conexão da sala cai repentinamente)
- **Sintoma:** A OpenAI gera o texto, mas o áudio não toca no frontend, ou o Worker crasha com `panic` na hora de escrever no track.
- **Causa (A):** Erros de API (Rate Limit 429, 401 Unauthorized, 404 Not Found devido a voz/modelo inexistente). **Sempre verifique os logs de erro da API.**
- **Causa (B):** Incompatibilidade de Codec. APIs de TTS geralmente retornam MP3 ou WAV. O LiveKit (WebRTC) exige envio em pacotes **Opus** formatados com duração exata (`20ms`).
- **Solução:** Usar `openai.SpeechResponseFormatOpus`. Decodificar o arquivo OGG recebido usando `github.com/pion/webrtc/v4/pkg/media/oggreader` e iterar por todas as páginas (`ParseNextPage`) calculando a duração pela diferença do `GranulePosition` antes de chamar `WriteSample`.

### 4. Cache Persistente do Navegador (Service Worker)
- **Sintoma:** Alterações no frontend não refletem no deploy de produção (Tela cinza `ERR_FAILED` ou código antigo no Console).
- **Causa:** O Service Worker interceptou o HTML principal (`/`) com estratégia *Cache-First* e gerou um loop de redirecionamento opaco, ou apenas travou o script antigo.
- **Solução:**
  1. Alterar a versão do `CACHE_NAME` em `sw.js`.
  2. Implementar estratégia **Network-First** para requests `mode === 'navigate'` ou headers `text/html`.
  3. Recomendar ao usuário que use `Ctrl + F5` ou navegue em janela anônima para testar.

## Ações a Tomar pelo Agent
1. Peça permissão para olhar os **logs do Railway** (`railway logs -s worker | Select-Object -Last 50`).
2. Se necessário, crie um script temporário em `cmd/diagnostic/main.go` para validar individualmente se as chaves de API do LiveKit, Deepgram e OpenAI possuem saldo/cotas.
3. Não presuma falha no frontend sem antes confirmar se o backend processou a resposta do LLM e do TTS sem retornar erro.