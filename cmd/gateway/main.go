package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pb "ai-translator/api/proto"
	"ai-translator/internal/config"
	"ai-translator/internal/gateway"
	"ai-translator/internal/logging"
	"ai-translator/internal/transport"
)

func main() {
	cfg := config.Load()
	logger := logging.New(cfg.LogLevel)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	asrConn, err := transport.NewGRPCClient(ctx, cfg.ASRAddress, logger)
	if err != nil {
		logger.Error("failed to connect to ASR service", "error", err)
		os.Exit(1)
	}
	defer asrConn.Close()

	translatorConn, err := transport.NewGRPCClient(ctx, cfg.TranslatorAddr, logger)
	if err != nil {
		logger.Error("failed to connect to Translator service", "error", err)
		os.Exit(1)
	}
	defer translatorConn.Close()

	ttsConn, err := transport.NewGRPCClient(ctx, cfg.TTSAddress, logger)
	if err != nil {
		logger.Error("failed to connect to TTS service", "error", err)
		os.Exit(1)
	}
	defer ttsConn.Close()

	asrClient := pb.NewASRServiceClient(asrConn.Conn())
	translatorClient := pb.NewTranslatorServiceClient(translatorConn.Conn())
	ttsClient := pb.NewTTSServiceClient(ttsConn.Conn())

	sessionManager := gateway.NewSessionManager(asrClient, translatorClient, ttsClient, logger)
	wsHandler := gateway.NewWebSocketHandler(sessionManager, logger)
	router := gateway.NewRouter(wsHandler, logger)

	addr := fmt.Sprintf(":%d", cfg.GatewayPort)
	server := transport.NewWSServer(addr, router.Handler(), logger)

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			cancel()
		}
	}()

	logger.Info("gateway started", "port", cfg.GatewayPort)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		logger.Info("shutdown signal received")
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logger.Error("shutdown error", "error", err)
	}

	logger.Info("gateway stopped")
}
