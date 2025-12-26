package translator

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
	logger *slog.Logger
}

func NewGeminiClient(ctx context.Context, apiKey string, logger *slog.Logger) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(0.3)
	model.SetTopP(0.8)
	model.SetTopK(40)
	model.SetMaxOutputTokens(256)

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(SystemPrompt),
		},
	}

	return &GeminiClient{
		client: client,
		model:  model,
		logger: logger,
	}, nil
}

func (g *GeminiClient) Close() error {
	return g.client.Close()
}

func (g *GeminiClient) Translate(ctx context.Context, text, sourceLang, targetLang string, conversationContext []string) (string, error) {
	prompt := BuildTranslationPrompt(text, sourceLang, targetLang, conversationContext)

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("gemini generation failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no translation generated")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result.WriteString(string(text))
		}
	}

	return strings.TrimSpace(result.String()), nil
}

func (g *GeminiClient) TranslateStream(ctx context.Context, text, sourceLang, targetLang string, conversationContext []string) (<-chan string, <-chan error) {
	textCh := make(chan string, 10)
	errCh := make(chan error, 1)

	go func() {
		defer close(textCh)
		defer close(errCh)

		prompt := BuildTranslationPrompt(text, sourceLang, targetLang, conversationContext)

		iter := g.model.GenerateContentStream(ctx, genai.Text(prompt))

		for {
			resp, err := iter.Next()
			if err != nil {
				if err.Error() != "iterator done" {
					errCh <- err
				}
				return
			}

			for _, cand := range resp.Candidates {
				if cand.Content == nil {
					continue
				}
				for _, part := range cand.Content.Parts {
					if text, ok := part.(genai.Text); ok {
						select {
						case textCh <- string(text):
						case <-ctx.Done():
							errCh <- ctx.Err()
							return
						}
					}
				}
			}
		}
	}()

	return textCh, errCh
}
