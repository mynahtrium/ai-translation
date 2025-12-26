package audio

import (
	"math"
)

type VAD struct {
	threshold     float64
	minSpeechMs   int
	minSilenceMs  int
	sampleRate    int
	speechSamples int
	silentSamples int
	isSpeaking    bool
}

func NewVAD(threshold float64, minSpeechMs, minSilenceMs int) *VAD {
	return &VAD{
		threshold:    threshold,
		minSpeechMs:  minSpeechMs,
		minSilenceMs: minSilenceMs,
		sampleRate:   SampleRate,
	}
}

func (v *VAD) Process(samples []int16) bool {
	energy := v.calculateEnergy(samples)
	isSpeech := energy > v.threshold

	if isSpeech {
		v.speechSamples += len(samples)
		v.silentSamples = 0
	} else {
		v.silentSamples += len(samples)
	}

	minSpeechSamples := SamplesForDuration(v.minSpeechMs)
	minSilentSamples := SamplesForDuration(v.minSilenceMs)

	if !v.isSpeaking && v.speechSamples >= minSpeechSamples {
		v.isSpeaking = true
	}

	if v.isSpeaking && v.silentSamples >= minSilentSamples {
		v.isSpeaking = false
		v.speechSamples = 0
	}

	return v.isSpeaking
}

func (v *VAD) calculateEnergy(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}

	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func (v *VAD) IsSpeaking() bool {
	return v.isSpeaking
}

func (v *VAD) Reset() {
	v.speechSamples = 0
	v.silentSamples = 0
	v.isSpeaking = false
}
