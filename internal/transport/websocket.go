package transport

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  16384,
	WriteBufferSize: 16384,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSConn struct {
	conn      *websocket.Conn
	mu        sync.Mutex
	logger    *slog.Logger
	sessionID string
	closed    bool
}

func UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, logger *slog.Logger, sessionID string) (*WSConn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &WSConn{
		conn:      conn,
		logger:    logger,
		sessionID: sessionID,
	}, nil
}

func (ws *WSConn) ReadMessage() (int, []byte, error) {
	return ws.conn.ReadMessage()
}

func (ws *WSConn) WriteMessage(messageType int, data []byte) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return websocket.ErrCloseSent
	}

	return ws.conn.WriteMessage(messageType, data)
}

func (ws *WSConn) WriteBinary(data []byte) error {
	return ws.WriteMessage(websocket.BinaryMessage, data)
}

func (ws *WSConn) WriteText(data string) error {
	return ws.WriteMessage(websocket.TextMessage, []byte(data))
}

func (ws *WSConn) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if ws.closed {
		return nil
	}

	ws.closed = true
	ws.conn.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second),
	)
	return ws.conn.Close()
}

func (ws *WSConn) SetReadDeadline(t time.Time) error {
	return ws.conn.SetReadDeadline(t)
}

func (ws *WSConn) SetWriteDeadline(t time.Time) error {
	return ws.conn.SetWriteDeadline(t)
}

func (ws *WSConn) SessionID() string {
	return ws.sessionID
}

type WSServer struct {
	server *http.Server
	logger *slog.Logger
}

func NewWSServer(addr string, handler http.Handler, logger *slog.Logger) *WSServer {
	return &WSServer{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		logger: logger,
	}
}

func (s *WSServer) Start() error {
	s.logger.Info("WebSocket server starting", "address", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *WSServer) Stop(ctx context.Context) error {
	s.logger.Info("WebSocket server stopping")
	return s.server.Shutdown(ctx)
}
