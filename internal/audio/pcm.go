package audio

const (
	SampleRate     = 16000
	Channels       = 1
	BitsPerSample  = 16
	BytesPerSample = BitsPerSample / 8
)

func PCMToBytes(samples []int16) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		buf[i*2] = byte(s)
		buf[i*2+1] = byte(s >> 8)
	}
	return buf
}

func BytesToPCM(data []byte) []int16 {
	samples := make([]int16, len(data)/2)
	for i := range samples {
		samples[i] = int16(data[i*2]) | int16(data[i*2+1])<<8
	}
	return samples
}

func DurationMs(samples int) int {
	return (samples * 1000) / SampleRate
}

func SamplesForDuration(ms int) int {
	return (ms * SampleRate) / 1000
}
