package transport

import (
	"context"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
	logger   *slog.Logger
}

func NewGRPCServer(logger *slog.Logger) *GRPCServer {
	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	return &GRPCServer{
		server: grpc.NewServer(opts...),
		logger: logger,
	}
}

func (s *GRPCServer) Server() *grpc.Server {
	return s.server
}

func (s *GRPCServer) Start(address string) error {
	var err error
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.logger.Info("gRPC server starting", "address", address)
	return s.server.Serve(s.listener)
}

func (s *GRPCServer) Stop(ctx context.Context) {
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
	case <-ctx.Done():
		s.server.Stop()
		s.logger.Warn("gRPC server forced stop")
	}
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	logger *slog.Logger
}

func NewGRPCClient(ctx context.Context, address string, logger *slog.Logger) (*GRPCClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *GRPCClient) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}
