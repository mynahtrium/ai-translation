package gateway

import (
	"log/slog"
	"net/http"
)

type Router struct {
	mux       *http.ServeMux
	wsHandler *WebSocketHandler
	logger    *slog.Logger
}

func NewRouter(wsHandler *WebSocketHandler, logger *slog.Logger) *Router {
	r := &Router{
		mux:       http.NewServeMux(),
		wsHandler: wsHandler,
		logger:    logger,
	}
	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	r.mux.HandleFunc("/ws", r.wsHandler.HandleConnection)
	r.mux.HandleFunc("/health", r.handleHealth)
	r.mux.HandleFunc("/ready", r.handleReady)
}

func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

func (r *Router) handleReady(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ready"}`))
}

func (r *Router) Handler() http.Handler {
	return r.mux
}
