package gateway

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"ai-translator/internal/transport"
	"ai-translator/internal/util"

	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	sessionManager *SessionManager
	logger         *slog.Logger
}

func NewWebSocketHandler(sm *SessionManager, logger *slog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		sessionManager: sm,
		logger:         logger,
	}
}

type ClientConfig struct {
	SourceLanguage string `json:"source_language"`
	TargetLanguage string `json:"target_language"`
}

func (h *WebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	sessionID := util.NewSessionID()
	logger := h.logger.With("session_id", sessionID)

	wsConn, err := transport.UpgradeToWebSocket(w, r, logger, sessionID)
	if err != nil {
		logger.Error("websocket upgrade failed", "error", err)
		return
	}

	session := h.sessionManager.Create(sessionID, wsConn, logger)

	ctx, cancel := context.WithCancel(r.Context())
	defer func() {
		cancel()
		h.sessionManager.Remove(sessionID)
		wsConn.Close()
		logger.Info("session ended")
	}()

	logger.Info("new session started")

	configReceived := false
	var config ClientConfig

	for {
		msgType, data, err := wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logger.Error("websocket read error", "error", err)
			}
			return
		}

		switch msgType {
		case websocket.TextMessage:
			if err := json.Unmarshal(data, &config); err != nil {
				logger.Warn("invalid config message", "error", err)
				continue
			}
			session.SetLanguages(config.SourceLanguage, config.TargetLanguage)
			configReceived = true
			logger.Info("config received", "source", config.SourceLanguage, "target", config.TargetLanguage)

			if err := wsConn.WriteText(`{"status":"ready"}`); err != nil {
				logger.Error("failed to send ready status", "error", err)
				return
			}

		case websocket.BinaryMessage:
			if !configReceived {
				logger.Warn("audio received before config")
				continue
			}

			if err := session.ProcessAudio(ctx, data); err != nil {
				logger.Error("audio processing error", "error", err)
			}
		}
	}
}
