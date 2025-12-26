# AI Translator

Real-time speech-to-speech AI translation system built with Go, gRPC, WebSockets, Google Cloud Speech-to-Text, Gemini Text API, and Google Cloud Text-to-Speech.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   Gateway   │────▶│     ASR     │────▶│ Google STT  │
│  (Browser)  │     │  (WebSocket)│     │   (gRPC)    │     │             │
└─────────────┘     └──────┬──────┘     └─────────────┘     └─────────────┘
       ▲                   │
       │                   ▼
       │            ┌─────────────┐     ┌─────────────┐
       │            │ Translator  │────▶│   Gemini    │
       │            │   (gRPC)    │     │     API     │
       │            └──────┬──────┘     └─────────────┘
       │                   │
       │                   ▼
       │            ┌─────────────┐     ┌─────────────┐
       └────────────│     TTS     │────▶│ Google TTS  │
                    │   (gRPC)    │     │             │
                    └─────────────┘     └─────────────┘
```

## Features

- **Real-time streaming**: Audio is processed as it arrives, no batch processing
- **Automatic language detection**: Detects source language automatically
- **N-way translation**: Supports translation between multiple languages
- **Low latency**: Partial ASR results are translated immediately
- **Context awareness**: Maintains conversation history for coherent translations

## Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Google Cloud account with Speech-to-Text and Text-to-Speech APIs enabled
- Gemini API key

## Setup

1. Clone and configure:
```bash
cp .env.example .env
# Edit .env with your credentials
```

2. Set up Google Cloud credentials:
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json
```

3. Generate proto files:
```bash
make proto
```

4. Run with Docker:
```bash
make docker-up
```

Or run locally:
```bash
# Terminal 1 - ASR
make run-asr

# Terminal 2 - Translator  
make run-translator

# Terminal 3 - TTS
make run-tts

# Terminal 4 - Gateway
make run-gateway
```

## WebSocket API

Connect to `ws://localhost:8080/ws`

### Protocol

1. Send configuration (JSON):
```json
{
  "source_language": "en-US",
  "target_language": "es-ES"
}
```

2. Receive ready confirmation:
```json
{"status": "ready"}
```

3. Send audio (binary): 16-bit PCM, 16kHz, mono

4. Receive translated audio (binary): 16-bit PCM, 16kHz, mono

## Supported Languages

- English (en-US, en-GB)
- Spanish (es-ES, es-MX)
- French (fr-FR)
- German (de-DE)
- Italian (it-IT)
- Portuguese (pt-BR, pt-PT)
- Japanese (ja-JP)
- Korean (ko-KR)
- Chinese (zh-CN, zh-TW)
- Russian (ru-RU)
- Arabic (ar-SA)
- Hindi (hi-IN)
- Turkish (tr-TR)

## Project Structure

```
ai-translator/
├── api/proto/           # gRPC service definitions
├── cmd/                 # Service entry points
│   ├── gateway/         # WebSocket gateway
│   ├── asr/             # Speech-to-text service
│   ├── translator/      # Translation service
│   └── tts/             # Text-to-speech service
├── internal/            # Internal packages
│   ├── audio/           # PCM handling, buffering, VAD
│   ├── asr/             # Google STT client
│   ├── translator/      # Gemini integration
│   ├── tts/             # Google TTS client
│   ├── gateway/         # WebSocket handling
│   ├── transport/       # gRPC/WS helpers
│   ├── config/          # Configuration
│   ├── logging/         # Structured logging
│   └── util/            # Utilities
├── pkg/sdk/             # Client SDKs
├── deployments/         # Docker configs
└── scripts/             # Dev scripts
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| GATEWAY_PORT | Gateway HTTP port | 8080 |
| ASR_PORT | ASR gRPC port | 50051 |
| TRANSLATOR_PORT | Translator gRPC port | 50052 |
| TTS_PORT | TTS gRPC port | 50053 |
| GEMINI_API_KEY | Gemini API key | - |
| GOOGLE_APPLICATION_CREDENTIALS | Path to GCP credentials | - |
| LOG_LEVEL | Logging level (debug/info/warn/error) | info |

## License

MIT