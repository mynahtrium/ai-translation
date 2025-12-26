package tts

import (
	"context"
	"log/slog"

	"ai-translator/internal/audio"
)

type StreamSynthesizer struct {
	client    *Client
	chunkSize int
	logger    *slog.Logger
}

func NewStreamSynthesizer(client *Client, logger *slog.Logger) *StreamSynthesizer {
	return &StreamSynthesizer{
		client:    client,
		chunkSize: audio.SamplesForDuration(100) * audio.BytesPerSample,
		logger:    logger,
	}
}

func (s *StreamSynthesizer) SynthesizeToChannel(ctx context.Context, text string, cfg SynthesizeConfig, out chan<- []byte) error {
	audioData, err := s.client.Synthesize(ctx, text, cfg)
	if err != nil {
		return err
	}

	return s.streamAudioData(ctx, audioData, out)
}

func (s *StreamSynthesizer) streamAudioData(ctx context.Context, data []byte, out chan<- []byte) error {
	for offset := 0; offset < len(data); offset += s.chunkSize {
		end := offset + s.chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[offset:end]

		select {
		case out <- chunk:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

type StreamResult struct {
	Audio []byte
	Error error
	Final bool
}

func (s *StreamSynthesizer) SynthesizeStreaming(ctx context.Context, textIn <-chan string, cfg SynthesizeConfig) <-chan StreamResult {
	results := make(chan StreamResult, 10)

	go func() {
		defer close(results)

		for {
			select {
			case <-ctx.Done():
				results <- StreamResult{Error: ctx.Err()}
				return
			case text, ok := <-textIn:
				if !ok {
					results <- StreamResult{Final: true}
					return
				}

				if text == "" {
					continue
				}

				audioData, err := s.client.Synthesize(ctx, text, cfg)
				if err != nil {
					results <- StreamResult{Error: err}
					continue
				}

				for offset := 0; offset < len(audioData); offset += s.chunkSize {
					end := offset + s.chunkSize
					if end > len(audioData) {
						end = len(audioData)
					}

					select {
					case results <- StreamResult{Audio: audioData[offset:end]}:
					case <-ctx.Done():
						results <- StreamResult{Error: ctx.Err()}
						return
					}
				}
			}
		}
	}()

	return results
}
