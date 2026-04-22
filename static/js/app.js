// Alpine.js component definition for the Interview Player
document.addEventListener('alpine:init', () => {
    Alpine.data('interviewPlayer', () => ({
        isConnected: false,
        isMuted: true,
        aiStatus: 'idle', // 'idle', 'listening', 'processing', 'speaking'
        aiStatusText: 'Aguardando início da sessão',
        candidateInstruction: 'Permita o acesso ao microfone e clique em iniciar.',
        
        room: null,
        audioElement: null,

        async init() {
            console.log('Nova Voice AI Player Initialized');
            
            // Create a hidden audio element to play the AI voice
            this.audioElement = document.createElement('audio');
            this.audioElement.autoplay = true;
            document.body.appendChild(this.audioElement);
            
            // Attempt to connect immediately (In a real app, you'd fetch the token first)
            await this.connectToLiveKit();
        },

        async connectToLiveKit() {
            try {
                this.candidateInstruction = 'Conectando ao servidor de entrevista...';
                
                // Initialize LiveKit Room
                this.room = new LivekitClient.Room({
                    adaptiveStream: true,
                    dynacast: true,
                });

                // Listeners for Room Events
                this.room.on(LivekitClient.RoomEvent.Connected, () => {
                    this.isConnected = true;
                    this.candidateInstruction = 'Você está conectado. Desmute o microfone para falar com a IA.';
                    this.setAiStatus('idle', 'A Inteligência Artificial está aguardando você falar.');
                });

                this.room.on(LivekitClient.RoomEvent.Disconnected, () => {
                    this.isConnected = false;
                    this.candidateInstruction = 'Conexão encerrada.';
                    this.setAiStatus('idle', 'Sessão encerrada.');
                });

                this.room.on(LivekitClient.RoomEvent.TrackSubscribed, (track, publication, participant) => {
                    if (track.kind === LivekitClient.Track.Kind.Audio) {
                        console.log('AI Audio Track subscribed');
                        track.attach(this.audioElement);
                        
                        // Fake state transitions based on track activity (in reality, driven by DataChannels or AudioAnalyzer)
                        this.setAiStatus('speaking', 'A Inteligência Artificial está falando.');
                        
                        // Listen for silence to go back to listening
                        track.on(LivekitClient.TrackEvent.Muted, () => {
                             this.setAiStatus('listening', 'A Inteligência Artificial está ouvindo.');
                        });
                        track.on(LivekitClient.TrackEvent.Unmuted, () => {
                             this.setAiStatus('speaking', 'A Inteligência Artificial está falando.');
                        });
                    }
                });

                // Read token from DOM element attribute
                const rootEl = document.querySelector('[data-token]');
                const token = rootEl ? rootEl.getAttribute('data-token') : '';
                
                if (!token) {
                    throw new Error("Token não encontrado");
                }
                
                // Em ambiente de produção real, você usaria o URL correto do seu LiveKit server
                const wsUrl = window.location.hostname.includes('127.0.0.1') ? "ws://127.0.0.1:7880" : "wss://seu-livekit-server.com";
                
                console.log("Token received, connecting to LiveKit at:", wsUrl);
                
                await this.room.connect(wsUrl, token);
                
                this.isConnected = true;
                this.candidateInstruction = 'Entrevista Pronta. Clique no microfone para começar.';
                this.setAiStatus('listening', 'A IA está pronta e ouvindo.');

            } catch (error) {
                console.error('Failed to connect to LiveKit', error);
                this.candidateInstruction = 'Erro ao conectar. Tente recarregar a página.';
            }
        },

        async toggleMute() {
            if (!this.isConnected) return;

            try {
                if (this.isMuted) {
                    // Turn Mic ON
                    await this.room.localParticipant.setMicrophoneEnabled(true);
                    this.isMuted = false;
                    this.candidateInstruction = 'Pode falar, a IA está te ouvindo.';
                    this.setAiStatus('listening', 'Microfone ativado. A IA está ouvindo.');
                } else {
                    // Turn Mic OFF
                    await this.room.localParticipant.setMicrophoneEnabled(false);
                    this.isMuted = true;
                    this.candidateInstruction = 'Microfone mutado.';
                    this.setAiStatus('idle', 'Microfone desativado.');
                }
            } catch (error) {
                console.error('Error toggling microphone', error);
                alert('Não foi possível acessar o microfone. Verifique as permissões do navegador.');
            }
        },

        async endInterview() {
            if (confirm('Tem certeza que deseja encerrar a entrevista?')) {
                if (this.room) {
                    this.room.disconnect();
                }
                this.isConnected = false;
                this.candidateInstruction = 'Entrevista encerrada com sucesso. Obrigado!';
                this.setAiStatus('idle', 'Entrevista finalizada.');
                
                // Optionally redirect to a feedback/thank you page
                setTimeout(() => {
                    window.location.href = '/';
                }, 3000);
            }
        },

        setAiStatus(status, accessibleText) {
            this.aiStatus = status;
            this.aiStatusText = accessibleText;
        }
    }));
});
