package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	pb "ai-translator/api/proto"
	"ai-translator/internal/audio"
	"ai-translator/internal/config"
	"ai-translator/internal/logging"
	"ai-translator/internal/transport"
	"ai-translator/internal/tts"
)

type ttsServer struct {
	pb.UnimplementedTTSServiceServer
	client *tts.Client
	logger *slog.Logger
}

func (s *ttsServer) Synthesize(req *pb.TTSRequest, stream pb.TTSService_SynthesizeServer) error {
	ctx := stream.Context()
	logger := s.logger.With("session_id", req.SessionId)

	langCode := tts.NormalizeLanguageForTTS(req.LanguageCode)
	cfg := tts.DefaultSynthesizeConfig(langCode)
	cfg.VoiceName = tts.GetVoiceForLanguage(langCode)

	if req.VoiceConfig != nil {
		if req.VoiceConfig.VoiceName != "" {
			cfg.VoiceName = req.VoiceConfig.VoiceName
		}
		if req.VoiceConfig.SpeakingRate > 0 {
			cfg.SpeakingRate = float64(req.VoiceConfig.SpeakingRate)
		}
		cfg.Pitch = float64(req.VoiceConfig.Pitch)
	}

	audioData, err := s.client.Synthesize(ctx, req.Text, cfg)
	if err != nil {
		logger.Error("synthesis failed", "error", err)
		return err
	}

	chunkSize := audio.SamplesForDuration(100) * audio.BytesPerSample

	for offset := 0; offset < len(audioData); offset += chunkSize {
		end := offset + chunkSize
		if end > len(audioData) {
			end = len(audioData)
		}

		chunk := audioData[offset:end]
		isFinal := end >= len(audioData)

		resp := &pb.TTSResponse{
			SessionId: req.SessionId,
			Audio: &pb.AudioChunk{
				Data:       chunk,
				SampleRate: audio.SampleRate,
				Channels:   audio.Channels,
			},
			IsFinal: isFinal,
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	logger.Debug("synthesis complete", "text_len", len(req.Text), "audio_len", len(audioData))
	return nil
}

func (s *ttsServer) StreamSynthesize(stream pb.TTSService_StreamSynthesizeServer) error {
	ctx := stream.Context()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		langCode := tts.NormalizeLanguageForTTS(req.LanguageCode)
		cfg := tts.DefaultSynthesizeConfig(langCode)
		cfg.VoiceName = tts.GetVoiceForLanguage(langCode)

		audioData, err := s.client.Synthesize(ctx, req.Text, cfg)
		if err != nil {
			s.logger.Error("synthesis failed", "error", err, "session_id", req.SessionId)
			continue
		}

		chunkSize := audio.SamplesForDuration(100) * audio.BytesPerSample

		for offset := 0; offset < len(audioData); offset += chunkSize {
			end := offset + chunkSize
			if end > len(audioData) {
				end = len(audioData)
			}

			resp := &pb.TTSResponse{
				SessionId: req.SessionId,
				Audio: &pb.AudioChunk{
					Data:       audioData[offset:end],
					SampleRate: audio.SampleRate,
					Channels:   audio.Channels,
				},
				IsFinal: end >= len(audioData),
			}

			if err := stream.Send(resp); err != nil {
				return err
			}
		}
	}
}

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ttsClient, err := tts.NewClient(ctx, logger)
	if err != nil {
		logger.Error("failed to create TTS client", "error", err)
		os.Exit(1)
	}
	defer ttsClient.Close()

	grpcServer := transport.NewGRPCServer(logger)
	pb.RegisterTTSServiceServer(grpcServer.Server(), &ttsServer{
		client: ttsClient,
		logger: logger,
	})

	go func() {
		addr := fmt.Sprintf(":%d", cfg.TTSPort)
		if err := grpcServer.Start(addr); err != nil {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	logger.Info("TTS service started", "port", cfg.TTSPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Info("shutdown signal received")
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	grpcServer.Stop(shutdownCtx)
	logger.Info("TTS service stopped")
}
