package sdk

import (
	"context"
	"io"

	pb "ai-translator/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ASRClient struct {
	conn   *grpc.ClientConn
	client pb.ASRServiceClient
}

func NewASRClient(ctx context.Context, address string) (*ASRClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ASRClient{
		conn:   conn,
		client: pb.NewASRServiceClient(conn),
	}, nil
}

func (c *ASRClient) Close() error {
	return c.conn.Close()
}

type ASRStream struct {
	stream pb.ASRService_StreamingRecognizeClient
}

func (c *ASRClient) StartStream(ctx context.Context, sessionID string, languages []string) (*ASRStream, error) {
	stream, err := c.client.StreamingRecognize(ctx)
	if err != nil {
		return nil, err
	}

	config := &pb.ASRRequest{
		Request: &pb.ASRRequest_Config{
			Config: &pb.StreamingConfig{
				Session: &pb.SessionInfo{
					SessionId: sessionID,
				},
				EnableAutomaticPunctuation: true,
				EnableLanguageDetection:    true,
				LanguageCodes:              languages,
			},
		},
	}

	if err := stream.Send(config); err != nil {
		return nil, err
	}

	return &ASRStream{stream: stream}, nil
}

func (s *ASRStream) SendAudio(data []byte, sampleRate, channels int32) error {
	return s.stream.Send(&pb.ASRRequest{
		Request: &pb.ASRRequest_Audio{
			Audio: &pb.AudioChunk{
				Data:       data,
				SampleRate: sampleRate,
				Channels:   channels,
			},
		},
	})
}

func (s *ASRStream) Recv() (*pb.ASRResponse, error) {
	return s.stream.Recv()
}

func (s *ASRStream) CloseSend() error {
	return s.stream.CloseSend()
}

type TranscriptCallback func(transcript string, isFinal bool, language string)

func (s *ASRStream) Process(ctx context.Context, cb TranscriptCallback) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		resp, err := s.stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		cb(resp.Transcript, resp.IsFinal, resp.DetectedLanguage)
	}
}
