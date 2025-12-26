package gateway

import (
	"context"
	"log/slog"
	"sync"

	pb "ai-translator/api/proto"
	"ai-translator/internal/audio"
	"ai-translator/internal/transport"
)

type Session struct {
	ID               string
	conn             *transport.WSConn
	logger           *slog.Logger
	audioBuffer      *audio.Buffer
	sourceLang       string
	targetLang       string
	mu               sync.RWMutex
	asrClient        pb.ASRServiceClient
	translatorClient pb.TranslatorServiceClient
	ttsClient        pb.TTSServiceClient
	audioChan        chan []byte
	closed           bool
}

type SessionManager struct {
	sessions         map[string]*Session
	mu               sync.RWMutex
	asrClient        pb.ASRServiceClient
	translatorClient pb.TranslatorServiceClient
	ttsClient        pb.TTSServiceClient
	logger           *slog.Logger
}

func NewSessionManager(asrClient pb.ASRServiceClient, translatorClient pb.TranslatorServiceClient, ttsClient pb.TTSServiceClient, logger *slog.Logger) *SessionManager {
	return &SessionManager{
		sessions:         make(map[string]*Session),
		asrClient:        asrClient,
		translatorClient: translatorClient,
		ttsClient:        ttsClient,
		logger:           logger,
	}
}

func (m *SessionManager) Create(id string, conn *transport.WSConn, logger *slog.Logger) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := &Session{
		ID:               id,
		conn:             conn,
		logger:           logger,
		audioBuffer:      audio.NewBuffer(16000 * 2 * 30),
		asrClient:        m.asrClient,
		translatorClient: m.translatorClient,
		ttsClient:        m.ttsClient,
		audioChan:        make(chan []byte, 100),
	}

	m.sessions[id] = session

	go session.startPipeline()

	return session
}

func (m *SessionManager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

func (m *SessionManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if session, ok := m.sessions[id]; ok {
		session.Close()
		delete(m.sessions, id)
	}
}

func (s *Session) SetLanguages(source, target string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sourceLang = source
	s.targetLang = target
}

func (s *Session) GetLanguages() (string, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sourceLang, s.targetLang
}

func (s *Session) ProcessAudio(ctx context.Context, data []byte) error {
	select {
	case s.audioChan <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.logger.Warn("audio channel full, dropping chunk")
		return nil
	}
}

func (s *Session) startPipeline() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	asrStream, err := s.asrClient.StreamingRecognize(ctx)
	if err != nil {
		s.logger.Error("failed to start ASR stream", "error", err)
		return
	}

	sourceLang, targetLang := s.GetLanguages()

	err = asrStream.Send(&pb.ASRRequest{
		Request: &pb.ASRRequest_Config{
			Config: &pb.StreamingConfig{
				Session: &pb.SessionInfo{
					SessionId: s.ID,
				},
				EnableAutomaticPunctuation: true,
				EnableLanguageDetection:    true,
				LanguageCodes:              []string{sourceLang},
			},
		},
	})
	if err != nil {
		s.logger.Error("failed to send ASR config", "error", err)
		return
	}

	transcriptChan := make(chan *pb.ASRResponse, 10)
	translatedChan := make(chan *pb.TranslateResponse, 10)
	audioChan := make(chan []byte, 100)

	go s.forwardAudioToASR(ctx, asrStream)
	go s.receiveASRResponses(ctx, asrStream, transcriptChan)
	go s.translateTranscripts(ctx, transcriptChan, translatedChan, targetLang)
	go s.synthesizeAndStream(ctx, translatedChan, audioChan, targetLang)
	go s.streamAudioToClient(ctx, audioChan)

	<-ctx.Done()
}

func (s *Session) forwardAudioToASR(ctx context.Context, stream pb.ASRService_StreamingRecognizeClient) {
	for {
		select {
		case <-ctx.Done():
			stream.CloseSend()
			return
		case audioData, ok := <-s.audioChan:
			if !ok {
				stream.CloseSend()
				return
			}

			err := stream.Send(&pb.ASRRequest{
				Request: &pb.ASRRequest_Audio{
					Audio: &pb.AudioChunk{
						Data:       audioData,
						SampleRate: audio.SampleRate,
						Channels:   audio.Channels,
					},
				},
			})
			if err != nil {
				s.logger.Error("failed to send audio to ASR", "error", err)
				return
			}
		}
	}
}

func (s *Session) receiveASRResponses(ctx context.Context, stream pb.ASRService_StreamingRecognizeClient, out chan<- *pb.ASRResponse) {
	defer close(out)

	for {
		resp, err := stream.Recv()
		if err != nil {
			if ctx.Err() == nil {
				s.logger.Error("ASR receive error", "error", err)
			}
			return
		}

		select {
		case out <- resp:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Session) translateTranscripts(ctx context.Context, in <-chan *pb.ASRResponse, out chan<- *pb.TranslateResponse, targetLang string) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-in:
			if !ok {
				return
			}

			if resp.Transcript == "" {
				continue
			}

			transResp, err := s.translatorClient.Translate(ctx, &pb.TranslateRequest{
				SessionId:      s.ID,
				Text:           resp.Transcript,
				SourceLanguage: resp.DetectedLanguage,
				TargetLanguage: targetLang,
				IsFinal:        resp.IsFinal,
			})
			if err != nil {
				s.logger.Error("translation error", "error", err)
				continue
			}

			select {
			case out <- transResp:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (s *Session) synthesizeAndStream(ctx context.Context, in <-chan *pb.TranslateResponse, out chan<- []byte, targetLang string) {
	defer close(out)

	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-in:
			if !ok {
				return
			}

			if resp.TranslatedText == "" {
				continue
			}

			ttsStream, err := s.ttsClient.Synthesize(ctx, &pb.TTSRequest{
				SessionId:    s.ID,
				Text:         resp.TranslatedText,
				LanguageCode: targetLang,
			})
			if err != nil {
				s.logger.Error("TTS synthesis error", "error", err)
				continue
			}

			for {
				ttsResp, err := ttsStream.Recv()
				if err != nil {
					break
				}

				if ttsResp.Audio != nil && len(ttsResp.Audio.Data) > 0 {
					select {
					case out <- ttsResp.Audio.Data:
					case <-ctx.Done():
						return
					}
				}

				if ttsResp.IsFinal {
					break
				}
			}
		}
	}
}

func (s *Session) streamAudioToClient(ctx context.Context, in <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case audioData, ok := <-in:
			if !ok {
				return
			}

			if err := s.conn.WriteBinary(audioData); err != nil {
				s.logger.Error("failed to send audio to client", "error", err)
				return
			}
		}
	}
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.closed = true
	close(s.audioChan)
	s.audioBuffer.Close()
}
