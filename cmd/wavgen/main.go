package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/generators"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gopxl/beep/v2/wav"
	"github.com/jwetzell/morse-go"
)

var flagWPM uint
var flagFrequency float64
var flagSampleRate int
var flagVolume float64
var flagOutput string

func init() {
	flag.UintVar(&flagWPM, "wpm", 15, "the WPM of the generated morse audio")
	flag.Float64Var(&flagFrequency, "frequency", 440, "the frequency of the generated morse audio")
	flag.IntVar(&flagSampleRate, "samplerate", 48000, "the sample rate of the generated morse audio")
	flag.Float64Var(&flagVolume, "volume", 0.8, "the volume of the generated morse audio")
	flag.StringVar(&flagOutput, "output", "", "the output WAV file for the generated morse audio")
}

func main() {
	flag.Parse()
	msg := flag.Arg(0)
	if msg == "" {
		slog.Error("No message provided")
		return
	}
	msg = strings.ToUpper(msg)
	wpm := flagWPM
	frequency := flagFrequency
	volume := flagVolume

	if volume < 0 || volume > 1 {
		slog.Error("Volume must be between 0 and 1", "volume", volume)
		return
	}

	sr := beep.SampleRate(flagSampleRate)

	unitDuration := morse.WPMToElementDuration(wpm)
	streamers := []beep.Streamer{}

	states, err := morse.IntlMorseCode.EncodeString(msg)
	if err != nil {
		slog.Error("Failed to encode message to Morse code", "error", err)
		return
	}
	for index, state := range states {
		if index > 0 && index < len(states)-1 && states[index-1] < 2 {
			streamers = append(streamers, generators.Silence(sr.N(unitDuration)))
		}

		switch state {
		case morse.StateDit:
			streamer, err := NewCWStreamer(unitDuration, frequency, volume, sr, time.Millisecond*4, time.Millisecond*4)
			if err != nil {
				slog.Error("Failed to create CW streamer for dit", "error", err)
				return
			}
			streamers = append(streamers, streamer)
		case morse.StateDah:
			streamer, err := NewCWStreamer(unitDuration*3, frequency, volume, sr, time.Millisecond*4, time.Millisecond*4)
			if err != nil {
				slog.Error("Failed to create CW streamer for dah", "error", err)
				return
			}
			streamers = append(streamers, streamer)
		case morse.StateEOC:
			streamers = append(streamers, generators.Silence(sr.N(unitDuration*3)))
		case morse.StateEOW:
			streamers = append(streamers, generators.Silence(sr.N(unitDuration*7)))
		default:
			continue
		}

	}

	if flagOutput != "" {
		outFile, err := os.Create(flagOutput)
		if err != nil {
			slog.Error("Failed to create output file", "error", err)
			return
		}
		defer outFile.Close()

		err = wav.Encode(outFile, beep.Seq(streamers...), beep.Format{
			SampleRate:  sr,
			NumChannels: 2,
			Precision:   2,
		})
		if err != nil {
			slog.Error("Failed to encode WAV", "error", err)
			return
		}
	} else {
		speaker.Init(sr, 4800)
		speaker.PlayAndWait(beep.Seq(streamers...))
	}

}

type CWStreamer struct {
	remains             int
	dt                  float64
	t                   float64
	inDelta             float64
	inSamplesRemaining  int
	inMultiplier        float64
	outDelta            float64
	outSamplesRemaining int
	outMultiplier       float64
	volume              float64
}

func NewCWStreamer(duration time.Duration, frequency float64, volume float64, sr beep.SampleRate, riseTime time.Duration, fallTime time.Duration) (beep.Streamer, error) {
	if sr <= 0 {
		return nil, fmt.Errorf("invalid sample rate")
	}

	if frequency <= 0 {
		return nil, fmt.Errorf("invalid frequency")
	}

	totalSamples := sr.N(duration)

	inSamples := sr.N(riseTime)
	inDelta := 1.0 / float64(inSamples)
	outSamples := sr.N(fallTime)
	outDelta := 1.0 / float64(outSamples)
	return &CWStreamer{
		remains:             totalSamples,
		dt:                  frequency / float64(sr),
		volume:              volume,
		inDelta:             inDelta,
		inSamplesRemaining:  inSamples,
		inMultiplier:        0,
		outDelta:            outDelta,
		outSamplesRemaining: outSamples,
		outMultiplier:       1.0,
	}, nil
}

func (cw *CWStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	if cw.remains <= 0 {
		return 0, false
	}
	toStream := min(len(samples), cw.remains)
	for i := range samples[:toStream] {
		v := math.Sin(cw.t * 2.0 * math.Pi)
		samples[i][0] = v * cw.inMultiplier * cw.outMultiplier * cw.volume
		samples[i][1] = v * cw.inMultiplier * cw.outMultiplier * cw.volume
		_, cw.t = math.Modf(cw.t + cw.dt)
		if cw.inSamplesRemaining > 0 {
			cw.inMultiplier += cw.inDelta
			cw.inSamplesRemaining--
		}
		if cw.outSamplesRemaining > 0 && cw.remains-(i+1) < cw.outSamplesRemaining {
			cw.outMultiplier -= cw.outDelta
			cw.outSamplesRemaining--
		}
	}
	cw.remains -= toStream
	return toStream, true
}

func (cw *CWStreamer) Err() error {
	return nil
}
