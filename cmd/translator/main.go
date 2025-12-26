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
	"ai-translator/internal/config"
	"ai-translator/internal/logging"
	"ai-translator/internal/translator"
	"ai-translator/internal/transport"
)

type translatorServer struct {
	pb.UnimplementedTranslatorServiceServer
	client *translator.GeminiClient
	ctxMgr *translator.ContextManager
	logger *slog.Logger
}

func (s *translatorServer) Translate(ctx context.Context, req *pb.TranslateRequest) (*pb.TranslateResponse, error) {
	logger := s.logger.With("session_id", req.SessionId)

	convCtx := s.ctxMgr.Get(req.SessionId)
	recentContext := convCtx.GetRecentOriginals()

	translated, err := s.client.Translate(ctx, req.Text, req.SourceLanguage, req.TargetLanguage, recentContext)
	if err != nil {
		logger.Error("translation failed", "error", err)
		return nil, err
	}

	if req.IsFinal {
		convCtx.Add(req.Text, translated, req.SourceLanguage, req.TargetLanguage)
	}

	logger.Debug("translated", "source", req.Text, "target", translated)

	return &pb.TranslateResponse{
		SessionId:      req.SessionId,
		TranslatedText: translated,
		SourceLanguage: req.SourceLanguage,
		TargetLanguage: req.TargetLanguage,
		IsFinal:        req.IsFinal,
	}, nil
}

func (s *translatorServer) StreamTranslate(stream pb.TranslatorService_StreamTranslateServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		resp, err := s.Translate(stream.Context(), req)
		if err != nil {
			return err
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	if cfg.GeminiAPIKey == "" {
		logger.Error("GEMINI_API_KEY is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	geminiClient, err := translator.NewGeminiClient(ctx, cfg.GeminiAPIKey, logger)
	if err != nil {
		logger.Error("failed to create Gemini client", "error", err)
		os.Exit(1)
	}
	defer geminiClient.Close()

	ctxMgr := translator.NewContextManager()

	grpcServer := transport.NewGRPCServer(logger)
	pb.RegisterTranslatorServiceServer(grpcServer.Server(), &translatorServer{
		client: geminiClient,
		ctxMgr: ctxMgr,
		logger: logger,
	})

	go func() {
		addr := fmt.Sprintf(":%d", cfg.TranslatorPort)
		if err := grpcServer.Start(addr); err != nil {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	logger.Info("Translator service started", "port", cfg.TranslatorPort)

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
	logger.Info("Translator service stopped")
}
