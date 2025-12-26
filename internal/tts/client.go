package tts

import (
	"context"
	"fmt"
	"log/slog"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	tspb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type Client struct {
	client *texttospeech.Client
	logger *slog.Logger
}

func NewClient(ctx context.Context, logger *slog.Logger) (*Client, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
	}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

type SynthesizeConfig struct {
	LanguageCode string
	VoiceName    string
	SpeakingRate float64
	Pitch        float64
	SampleRate   int32
}

func DefaultSynthesizeConfig(languageCode string) SynthesizeConfig {
	return SynthesizeConfig{
		LanguageCode: languageCode,
		VoiceName:    "",
		SpeakingRate: 1.0,
		Pitch:        0.0,
		SampleRate:   16000,
	}
}

func (c *Client) Synthesize(ctx context.Context, text string, cfg SynthesizeConfig) ([]byte, error) {
	voice := &tspb.VoiceSelectionParams{
		LanguageCode: cfg.LanguageCode,
		SsmlGender:   tspb.SsmlVoiceGender_NEUTRAL,
	}

	if cfg.VoiceName != "" {
		voice.Name = cfg.VoiceName
	}

	req := &tspb.SynthesizeSpeechRequest{
		Input: &tspb.SynthesisInput{
			InputSource: &tspb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: voice,
		AudioConfig: &tspb.AudioConfig{
			AudioEncoding:   tspb.AudioEncoding_LINEAR16,
			SampleRateHertz: cfg.SampleRate,
			SpeakingRate:    cfg.SpeakingRate,
			Pitch:           cfg.Pitch,
		},
	}

	resp, err := c.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TTS synthesis failed: %w", err)
	}

	return resp.AudioContent, nil
}

func (c *Client) SynthesizeSSML(ctx context.Context, ssml string, cfg SynthesizeConfig) ([]byte, error) {
	voice := &tspb.VoiceSelectionParams{
		LanguageCode: cfg.LanguageCode,
		SsmlGender:   tspb.SsmlVoiceGender_NEUTRAL,
	}

	if cfg.VoiceName != "" {
		voice.Name = cfg.VoiceName
	}

	req := &tspb.SynthesizeSpeechRequest{
		Input: &tspb.SynthesisInput{
			InputSource: &tspb.SynthesisInput_Ssml{
				Ssml: ssml,
			},
		},
		Voice: voice,
		AudioConfig: &tspb.AudioConfig{
			AudioEncoding:   tspb.AudioEncoding_LINEAR16,
			SampleRateHertz: cfg.SampleRate,
			SpeakingRate:    cfg.SpeakingRate,
			Pitch:           cfg.Pitch,
		},
	}

	resp, err := c.client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TTS SSML synthesis failed: %w", err)
	}

	return resp.AudioContent, nil
}
