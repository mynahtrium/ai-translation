package asr

import (
	"context"
	"io"
	"log/slog"

	pb "ai-translator/api/proto"
)

type StreamHandler struct {
	client *Client
	logger *slog.Logger
}

func NewStreamHandler(client *Client, logger *slog.Logger) *StreamHandler {
	return &StreamHandler{
		client: client,
		logger: logger,
	}
}

func (h *StreamHandler) Handle(ctx context.Context, audioIn <-chan []byte, resultsOut chan<- *pb.ASRResponse, sessionID string) error {
	stream, err := h.client.StreamingRecognize(ctx)
	if err != nil {
		return err
	}

	cfg := DefaultStreamConfig()
	if err := stream.SendConfig(cfg); err != nil {
		return err
	}

	errCh := make(chan error, 2)

	go func() {
		for {
			select {
			case <-ctx.Done():
				stream.CloseSend()
				return
			case audio, ok := <-audioIn:
				if !ok {
					stream.CloseSend()
					return
				}
				if err := stream.SendAudio(audio); err != nil {
					if err != io.EOF {
						errCh <- err
					}
					return
				}
			}
		}
	}()

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				errCh <- err
				return
			}

			for _, result := range resp.Results {
				if len(result.Alternatives) == 0 {
					continue
				}

				alt := result.Alternatives[0]
				pbResp := &pb.ASRResponse{
					SessionId:        sessionID,
					Transcript:       alt.Transcript,
					IsFinal:          result.IsFinal,
					Stability:        result.Stability,
					DetectedLanguage: result.LanguageCode,
				}

				select {
				case resultsOut <- pbResp:
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
