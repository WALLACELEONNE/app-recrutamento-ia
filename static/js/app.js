// Alpine.js component definition for the Interview Player
document.addEventListener('alpine:init', () => {
    Alpine.data('interviewPlayer', () => {
        // Declared outside the returned object to avoid Alpine proxying
        // Proxies break complex WebRTC objects like LivekitClient.Room causing DataCloneError
        let room = null;
        let audioElement = null;
        let aiTrackTimeoutId = null;

        return {
            interviewStarted: false,
            isConnected: false,
            isMuted: true,
            isSpeaking: false,
            hasAiTrack: false,
            aiStatus: 'idle', // 'idle', 'listening', 'processing', 'speaking'
            aiStatusText: 'Aguardando início da sessão',
            candidateInstruction: 'Clique em "Iniciar Entrevista" e permita o acesso ao microfone.',
            timeRemaining: 15 * 60, // 15 minutos em segundos
            timerInterval: null,

            get formattedTime() {
                const minutes = Math.floor(this.timeRemaining / 60);
                const seconds = this.timeRemaining % 60;
                return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
            },

            async init() {
                console.log('Nova Voice AI Player Initialized');
            },

            async startInterview() {
                this.interviewStarted = true;
                
                // Create a hidden audio element to play the AI voice AFTER user interaction
                audioElement = document.createElement('audio');
                audioElement.autoplay = true;
                audioElement.playsInline = true;
                document.body.appendChild(audioElement);

                await this.connectToLiveKit();
            },

            async connectToLiveKit() {
                try {
                    this.candidateInstruction = 'Conectando ao servidor de entrevista...';
                    
                    // Initialize LiveKit Room
                    room = new LivekitClient.Room({
                        adaptiveStream: true,
                        dynacast: true,
                    });

                    // Listeners for Room Events
                    room.on(LivekitClient.RoomEvent.Connected, () => {
                        this.isConnected = true;
                        this.hasAiTrack = false;
                        this.candidateInstruction = 'Conectado. Aguardando a IA entrar na sala...';
                        this.setAiStatus('idle', 'A Inteligência Artificial está aguardando você falar.');
                        this.startTimer();
                        
                        // Garante que o contexto de áudio do navegador foi liberado
                        room.startAudio().catch(console.error);
                        
                        // Habilita o microfone automaticamente para disparar a introdução da IA
                        room.localParticipant.setMicrophoneEnabled(true).then(() => {
                            this.isMuted = false;
                        }).catch(e => {
                            console.error('Falha ao habilitar microfone automaticamente', e);
                            this.candidateInstruction = 'Por favor, clique em "Microfone Mutado" para habilitar o áudio.';
                        });

                        clearTimeout(aiTrackTimeoutId);
                        aiTrackTimeoutId = setTimeout(() => {
                            if (!this.hasAiTrack) {
                                console.error('AI worker did not publish an audio track in time');
                                this.candidateInstruction = 'A IA nao entrou na sala a tempo. Recarregue a pagina para reenfileirar o worker.';
                                this.setAiStatus('idle', 'A IA nao ficou disponivel.');
                            }
                        }, 12000);
                    });

                    room.on(LivekitClient.RoomEvent.Disconnected, () => {
                        this.isConnected = false;
                        this.hasAiTrack = false;
                        this.candidateInstruction = 'Conexão encerrada.';
                        this.setAiStatus('idle', 'Sessão encerrada.');
                        this.stopTimer();
                        clearTimeout(aiTrackTimeoutId);
                    });

                    room.on(LivekitClient.RoomEvent.ActiveSpeakersChanged, (speakers) => {
                        let localSpeaking = false;
                        let remoteSpeaking = false;

                        speakers.forEach((p) => {
                            if (p === room.localParticipant) {
                                localSpeaking = true;
                            } else {
                                remoteSpeaking = true;
                            }
                        });

                        this.isSpeaking = localSpeaking;
                        
                        if (remoteSpeaking) {
                            this.setAiStatus('speaking', 'A Inteligência Artificial está falando.');
                        } else {
                            if (!this.isMuted) {
                                this.setAiStatus('listening', 'A Inteligência Artificial está ouvindo.');
                            } else {
                                this.setAiStatus('idle', 'Aguardando ação do candidato.');
                            }
                        }
                    });

                    room.on(LivekitClient.RoomEvent.TrackSubscribed, (track, publication, participant) => {
                        if (track.kind === LivekitClient.Track.Kind.Audio) {
                            console.log('AI Audio Track subscribed');
                            this.hasAiTrack = true;
                            clearTimeout(aiTrackTimeoutId);
                            track.attach(audioElement);
                            audioElement.play().catch((error) => {
                                console.error('Failed to play AI audio track', error);
                            });
                            this.candidateInstruction = 'IA conectada. A entrevista comecara automaticamente.';
                            
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

                    // Read token and url from DOM element attribute
                    const rootEl = document.querySelector('[data-token]');
                    const token = rootEl ? rootEl.getAttribute('data-token') : '';
                    const wsUrl = rootEl ? rootEl.getAttribute('data-url') : '';
                    
                    if (!token) {
                        throw new Error("Token não encontrado");
                    }
                    
                    if (!wsUrl) {
                        throw new Error("URL do servidor não encontrada");
                    }
                    
                    console.log("Token received, connecting to LiveKit at:", wsUrl);
                    
                    await room.connect(wsUrl, token);
                    
                    this.isConnected = true;
                    this.candidateInstruction = 'Conectado ao LiveKit. Aguardando a presenca da IA...';
                    this.setAiStatus('idle', 'Conectado. Aguardando a IA entrar na sala.');

                } catch (error) {
                    console.error('Failed to connect to LiveKit', error);
                    this.candidateInstruction = 'Erro ao conectar. Tente recarregar a página.';
                }
            },

            async toggleMute() {
                if (!this.isConnected || !room) return;

                try {
                    if (this.isMuted) {
                        // Turn Mic ON
                        await room.localParticipant.setMicrophoneEnabled(true);
                        this.isMuted = false;
                        this.candidateInstruction = 'Pode falar, a IA está te ouvindo.';
                        this.setAiStatus('listening', 'Microfone ativado. A IA está ouvindo.');
                    } else {
                        // Turn Mic OFF
                        await room.localParticipant.setMicrophoneEnabled(false);
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
                    if (room) {
                        room.disconnect();
                    }
                    clearTimeout(aiTrackTimeoutId);
                    this.isConnected = false;
                    this.hasAiTrack = false;
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
            },

            startTimer() {
                if (this.timerInterval) clearInterval(this.timerInterval);
                this.timerInterval = setInterval(() => {
                    if (this.timeRemaining > 0) {
                        this.timeRemaining--;
                    } else {
                        this.stopTimer();
                        this.endInterview();
                        alert('O tempo da entrevista (15 minutos) expirou.');
                    }
                }, 1000);
            },

            stopTimer() {
                if (this.timerInterval) {
                    clearInterval(this.timerInterval);
                    this.timerInterval = null;
                }
            }
        };
    });
});
