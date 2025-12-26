package tts

var voiceMap = map[string]string{
	"en-US": "en-US-Neural2-J",
	"en-GB": "en-GB-Neural2-B",
	"es-ES": "es-ES-Neural2-B",
	"es-MX": "es-US-Neural2-B",
	"fr-FR": "fr-FR-Neural2-B",
	"de-DE": "de-DE-Neural2-B",
	"it-IT": "it-IT-Neural2-A",
	"pt-BR": "pt-BR-Neural2-B",
	"pt-PT": "pt-PT-Neural2-B",
	"ja-JP": "ja-JP-Neural2-C",
	"ko-KR": "ko-KR-Neural2-B",
	"zh-CN": "cmn-CN-Neural2-B",
	"zh-TW": "cmn-TW-Neural2-B",
	"ru-RU": "ru-RU-Neural2-B",
	"ar-SA": "ar-XA-Neural2-B",
	"hi-IN": "hi-IN-Neural2-B",
	"tr-TR": "tr-TR-Neural2-B",
	"nl-NL": "nl-NL-Neural2-B",
	"pl-PL": "pl-PL-Neural2-B",
	"sv-SE": "sv-SE-Neural2-A",
}

func GetVoiceForLanguage(languageCode string) string {
	if voice, ok := voiceMap[languageCode]; ok {
		return voice
	}
	return ""
}

func GetLanguageCodeFromVoice(voiceName string) string {
	for lang, voice := range voiceMap {
		if voice == voiceName {
			return lang
		}
	}
	return ""
}

func NormalizeLanguageForTTS(code string) string {
	mapping := map[string]string{
		"en": "en-US",
		"es": "es-ES",
		"fr": "fr-FR",
		"de": "de-DE",
		"it": "it-IT",
		"pt": "pt-BR",
		"ja": "ja-JP",
		"ko": "ko-KR",
		"zh": "zh-CN",
		"ru": "ru-RU",
		"ar": "ar-SA",
		"hi": "hi-IN",
		"tr": "tr-TR",
		"nl": "nl-NL",
		"pl": "pl-PL",
		"sv": "sv-SE",
	}

	if mapped, ok := mapping[code]; ok {
		return mapped
	}
	return code
}

func GetSupportedVoices() map[string]string {
	result := make(map[string]string)
	for k, v := range voiceMap {
		result[k] = v
	}
	return result
}
