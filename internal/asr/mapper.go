package asr

var languageMap = map[string]string{
	"en":    "en-US",
	"en-us": "en-US",
	"en-gb": "en-GB",
	"es":    "es-ES",
	"es-es": "es-ES",
	"es-mx": "es-MX",
	"fr":    "fr-FR",
	"fr-fr": "fr-FR",
	"de":    "de-DE",
	"de-de": "de-DE",
	"it":    "it-IT",
	"it-it": "it-IT",
	"pt":    "pt-BR",
	"pt-br": "pt-BR",
	"pt-pt": "pt-PT",
	"ja":    "ja-JP",
	"ja-jp": "ja-JP",
	"ko":    "ko-KR",
	"ko-kr": "ko-KR",
	"zh":    "zh-CN",
	"zh-cn": "zh-CN",
	"zh-tw": "zh-TW",
	"ru":    "ru-RU",
	"ru-ru": "ru-RU",
	"ar":    "ar-SA",
	"ar-sa": "ar-SA",
	"hi":    "hi-IN",
	"hi-in": "hi-IN",
	"tr":    "tr-TR",
	"tr-tr": "tr-TR",
}

func NormalizeLanguageCode(code string) string {
	if mapped, ok := languageMap[code]; ok {
		return mapped
	}
	return code
}

func GetSupportedLanguages() []string {
	return []string{
		"en-US", "en-GB", "es-ES", "es-MX", "fr-FR",
		"de-DE", "it-IT", "pt-BR", "pt-PT", "ja-JP",
		"ko-KR", "zh-CN", "zh-TW", "ru-RU", "ar-SA",
		"hi-IN", "tr-TR",
	}
}

func ExtractPrimaryLanguage(code string) string {
	if len(code) >= 2 {
		return code[:2]
	}
	return code
}
