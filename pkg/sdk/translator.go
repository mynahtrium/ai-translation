package sdk

import (
	"context"

	pb "ai-translator/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TranslatorClient struct {
	conn   *grpc.ClientConn
	client pb.TranslatorServiceClient
}

func NewTranslatorClient(ctx context.Context, address string) (*TranslatorClient, error) {
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &TranslatorClient{
		conn:   conn,
		client: pb.NewTranslatorServiceClient(conn),
	}, nil
}

func (c *TranslatorClient) Close() error {
	return c.conn.Close()
}

func (c *TranslatorClient) Translate(ctx context.Context, sessionID, text, sourceLang, targetLang string, isFinal bool) (string, error) {
	resp, err := c.client.Translate(ctx, &pb.TranslateRequest{
		SessionId:      sessionID,
		Text:           text,
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		IsFinal:        isFinal,
	})
	if err != nil {
		return "", err
	}

	return resp.TranslatedText, nil
}

type TranslateStream struct {
	stream pb.TranslatorService_StreamTranslateClient
}

func (c *TranslatorClient) StartStream(ctx context.Context) (*TranslateStream, error) {
	stream, err := c.client.StreamTranslate(ctx)
	if err != nil {
		return nil, err
	}

	return &TranslateStream{stream: stream}, nil
}

func (s *TranslateStream) Send(sessionID, text, sourceLang, targetLang string, isFinal bool) error {
	return s.stream.Send(&pb.TranslateRequest{
		SessionId:      sessionID,
		Text:           text,
		SourceLanguage: sourceLang,
		TargetLanguage: targetLang,
		IsFinal:        isFinal,
	})
}

func (s *TranslateStream) Recv() (*pb.TranslateResponse, error) {
	return s.stream.Recv()
}

func (s *TranslateStream) CloseSend() error {
	return s.stream.CloseSend()
}
