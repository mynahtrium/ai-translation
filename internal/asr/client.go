package asr

import (
	"context"
	"io"
	"log/slog"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "cloud.google.com/go/speech/apiv1/speechpb"
)

type Client struct {
	client *speech.Client
	logger *slog.Logger
}

func NewClient(ctx context.Context, logger *slog.Logger) (*Client, error) {
	client, err := speech.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) StreamingRecognize(ctx context.Context) (*Stream, error) {
	stream, err := c.client.StreamingRecognize(ctx)
	if err != nil {
		return nil, err
	}

	return &Stream{
		stream: stream,
		logger: c.logger,
	}, nil
}

type StreamConfig struct {
	SampleRate          int
	LanguageCodes       []string
	EnableAutoDetection bool
	EnablePunctuation   bool
	InterimResults      bool
}

func DefaultStreamConfig() StreamConfig {
	return StreamConfig{
		SampleRate:          16000,
		LanguageCodes:       []string{"en-US", "es-ES", "fr-FR", "de-DE", "tr-TR", "ja-JP", "zh-CN"},
		EnableAutoDetection: true,
		EnablePunctuation:   true,
		InterimResults:      true,
	}
}

type Stream struct {
	stream speechpb.Speech_StreamingRecognizeClient
	logger *slog.Logger
}

func (s *Stream) SendConfig(cfg StreamConfig) error {
	config := &speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					Encoding:                   speechpb.RecognitionConfig_LINEAR16,
					SampleRateHertz:            int32(cfg.SampleRate),
					LanguageCode:               cfg.LanguageCodes[0],
					AlternativeLanguageCodes:   cfg.LanguageCodes[1:],
					EnableAutomaticPunctuation: cfg.EnablePunctuation,
				},
				InterimResults: cfg.InterimResults,
			},
		},
	}

	return s.stream.Send(config)
}

func (s *Stream) SendAudio(data []byte) error {
	return s.stream.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_AudioContent{
			AudioContent: data,
		},
	})
}

func (s *Stream) Recv() (*speechpb.StreamingRecognizeResponse, error) {
	return s.stream.Recv()
}

func (s *Stream) CloseSend() error {
	return s.stream.CloseSend()
}

type RecognitionResult struct {
	Transcript       string
	IsFinal          bool
	Stability        float32
	DetectedLanguage string
}

func (s *Stream) ProcessResponses(ctx context.Context, results chan<- RecognitionResult) error {
	defer close(results)

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

		for _, result := range resp.Results {
			if len(result.Alternatives) == 0 {
				continue
			}

			alt := result.Alternatives[0]
			lang := result.LanguageCode
			if lang == "" && len(resp.Results) > 0 {
				lang = resp.Results[0].LanguageCode
			}

			select {
			case results <- RecognitionResult{
				Transcript:       alt.Transcript,
				IsFinal:          result.IsFinal,
				Stability:        result.Stability,
				DetectedLanguage: lang,
			}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}
