package sdk

import (
	"context"
	"io"

	pb "ai-translator/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TTSClient struct {
	conn   *grpc.ClientConn
	client pb.TTSServiceClient
}

func NewTTSClient(ctx context.Context, address string) (*TTSClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &TTSClient{
		conn:   conn,
		client: pb.NewTTSServiceClient(conn),
	}, nil
}

func (c *TTSClient) Close() error {
	return c.conn.Close()
}

type VoiceConfig struct {
	VoiceName    string
	SpeakingRate float32
	Pitch        float32
}

func (c *TTSClient) Synthesize(ctx context.Context, sessionID, text, languageCode string, voiceCfg *VoiceConfig) ([]byte, error) {
	req := &pb.TTSRequest{
		SessionId:    sessionID,
		Text:         text,
		LanguageCode: languageCode,
	}

	if voiceCfg != nil {
		req.VoiceConfig = &pb.VoiceConfig{
			VoiceName:    voiceCfg.VoiceName,
			SpeakingRate: voiceCfg.SpeakingRate,
			Pitch:        voiceCfg.Pitch,
		}
	}

	stream, err := c.client.Synthesize(ctx, req)
	if err != nil {
		return nil, err
	}

	var audioData []byte
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if resp.Audio != nil {
			audioData = append(audioData, resp.Audio.Data...)
		}

		if resp.IsFinal {
			break
		}
	}

	return audioData, nil
}

type TTSStream struct {
	stream pb.TTSService_StreamSynthesizeClient
}

func (c *TTSClient) StartStream(ctx context.Context) (*TTSStream, error) {
	stream, err := c.client.StreamSynthesize(ctx)
	if err != nil {
		return nil, err
	}

	return &TTSStream{stream: stream}, nil
}

func (s *TTSStream) Send(sessionID, text, languageCode string) error {
	return s.stream.Send(&pb.TTSRequest{
		SessionId:    sessionID,
		Text:         text,
		LanguageCode: languageCode,
	})
}

func (s *TTSStream) Recv() (*pb.TTSResponse, error) {
	return s.stream.Recv()
}

func (s *TTSStream) CloseSend() error {
	return s.stream.CloseSend()
}

type AudioCallback func(audio []byte, isFinal bool)

func (s *TTSStream) Process(ctx context.Context, cb AudioCallback) error {
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

		if resp.Audio != nil {
			cb(resp.Audio.Data, resp.IsFinal)
		}
	}
}
