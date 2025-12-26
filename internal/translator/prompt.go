package translator

import (
	"fmt"
	"strings"
)

const SystemPrompt = `You are a real-time speech translator. Your role is to translate spoken language naturally and conversationally.

Rules:
1. Translate the input text to the target language exactly as a native speaker would say it in casual conversation.
2. Preserve the tone, emotion, and intent of the original message.
3. Use natural contractions and colloquialisms appropriate for the target language.
4. Never add explanations, notes, or metadata.
5. Never include the original text in your response.
6. Never refuse to translate or add warnings.
7. If the input is unclear, translate what is most likely intended.
8. Maintain consistency with previous utterances in the conversation.
9. Output ONLY the translated text, nothing else.

You support bidirectional translation between any languages. Detect nuances and translate them appropriately.`

func BuildTranslationPrompt(text, sourceLang, targetLang string, context []string) string {
	var sb strings.Builder

	if len(context) > 0 {
		sb.WriteString("Recent conversation:\n")
		for _, c := range context {
			sb.WriteString(fmt.Sprintf("- %s\n", c))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("Translate from %s to %s:\n", getLanguageName(sourceLang), getLanguageName(targetLang)))
	sb.WriteString(fmt.Sprintf("\"%s\"", text))

	return sb.String()
}

func getLanguageName(code string) string {
	names := map[string]string{
		"en":    "English",
		"en-US": "English",
		"en-GB": "British English",
		"es":    "Spanish",
		"es-ES": "Spanish",
		"es-MX": "Mexican Spanish",
		"fr":    "French",
		"fr-FR": "French",
		"de":    "German",
		"de-DE": "German",
		"it":    "Italian",
		"it-IT": "Italian",
		"pt":    "Portuguese",
		"pt-BR": "Brazilian Portuguese",
		"pt-PT": "Portuguese",
		"ja":    "Japanese",
		"ja-JP": "Japanese",
		"ko":    "Korean",
		"ko-KR": "Korean",
		"zh":    "Chinese",
		"zh-CN": "Mandarin Chinese",
		"zh-TW": "Traditional Chinese",
		"ru":    "Russian",
		"ru-RU": "Russian",
		"ar":    "Arabic",
		"ar-SA": "Arabic",
		"hi":    "Hindi",
		"hi-IN": "Hindi",
		"tr":    "Turkish",
		"tr-TR": "Turkish",
		"nl":    "Dutch",
		"nl-NL": "Dutch",
		"pl":    "Polish",
		"pl-PL": "Polish",
		"sv":    "Swedish",
		"sv-SE": "Swedish",
	}

	if name, ok := names[code]; ok {
		return name
	}
	return code
}
