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
	internalASR "ai-translator/internal/asr"
	"ai-translator/internal/config"
	"ai-translator/internal/logging"
	"ai-translator/internal/transport"
)

type asrServer struct {
	pb.UnimplementedASRServiceServer
	client *internalASR.Client
	logger *slog.Logger
}

func (s *asrServer) StreamingRecognize(stream pb.ASRService_StreamingRecognizeServer) error {
	ctx := stream.Context()

	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	cfg, ok := firstMsg.Request.(*pb.ASRRequest_Config)
	if !ok {
		return fmt.Errorf("first message must be config")
	}

	sessionID := cfg.Config.Session.SessionId
	logger := s.logger.With("session_id", sessionID)
	logger.Info("ASR stream started")

	speechStream, err := s.client.StreamingRecognize(ctx)
	if err != nil {
		return err
	}

	streamCfg := internalASR.DefaultStreamConfig()
	if len(cfg.Config.LanguageCodes) > 0 {
		streamCfg.LanguageCodes = cfg.Config.LanguageCodes
	}
	streamCfg.EnablePunctuation = cfg.Config.EnableAutomaticPunctuation
	streamCfg.EnableAutoDetection = cfg.Config.EnableLanguageDetection

	if err := speechStream.SendConfig(streamCfg); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				speechStream.CloseSend()
				return
			}
			if err != nil {
				errCh <- err
				return
			}

			audio, ok := msg.Request.(*pb.ASRRequest_Audio)
			if !ok {
				continue
			}

			if err := speechStream.SendAudio(audio.Audio.Data); err != nil {
				if err != io.EOF {
					errCh <- err
				}
				return
			}
		}
	}()

	go func() {
		results := make(chan internalASR.RecognitionResult, 10)
		go func() {
			if err := speechStream.ProcessResponses(ctx, results); err != nil {
				errCh <- err
			}
		}()

		for result := range results {
			resp := &pb.ASRResponse{
				SessionId:        sessionID,
				Transcript:       result.Transcript,
				IsFinal:          result.IsFinal,
				Stability:        result.Stability,
				DetectedLanguage: result.DetectedLanguage,
			}

			if err := stream.Send(resp); err != nil {
				errCh <- err
				return
			}

			if result.IsFinal {
				logger.Debug("final transcript", "text", result.Transcript)
			}
		}
		errCh <- nil
	}()

	return <-errCh
}

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	asrClient, err := internalASR.NewClient(ctx, logger)
	if err != nil {
		logger.Error("failed to create ASR client", "error", err)
		os.Exit(1)
	}
	defer asrClient.Close()

	grpcServer := transport.NewGRPCServer(logger)
	pb.RegisterASRServiceServer(grpcServer.Server(), &asrServer{
		client: asrClient,
		logger: logger,
	})

	go func() {
		addr := fmt.Sprintf(":%d", cfg.ASRPort)
		if err := grpcServer.Start(addr); err != nil {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	logger.Info("ASR service started", "port", cfg.ASRPort)

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
	logger.Info("ASR service stopped")
}
