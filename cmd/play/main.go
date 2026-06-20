package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jwetzell/morse-go"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var (
	midiOut string
	wpm     uint
	ewpm    uint
)

func init() {
	flag.StringVar(&midiOut, "midi-out", "", "MIDI output device name")
	flag.UintVar(&wpm, "wpm", 20, "Words per minute for Morse code timing")
	flag.UintVar(&ewpm, "ewpm", 0, "Effective words per minute for Morse code timing")
}

func main() {
	flag.Parse()
	msg := flag.Arg(0)
	defer midi.CloseDriver()

	if msg == "" {
		slog.Error("No message provided to play")
		return
	}

	if midiOut == "" {
		slog.Error("No MIDI output device specified, using default")
		return
	}

	if wpm == 0 {
		slog.Error("WPM must be greater than 0")
		return
	}

	if ewpm == 0 {
		ewpm = wpm
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	midiOut, err := midi.FindOutPort(midiOut)
	if err != nil {
		slog.Error("MIDI output device not found", "device", midiOut)
		return
	}

	err = midiOut.Open()
	if err != nil {
		slog.Error("Failed to open MIDI output device", "device", midiOut, "error", err)
		return
	}
	defer midiOut.Close()
	fmt.Printf("\"%s\" -> %s\n", msg, midiOut)

	msg = strings.ToUpper(msg)

	interElementDuration := morse.WPMToElementDuration(ewpm)

	defer func() {
		<-ctx.Done()
		midiOut.Send([]byte{0x80, 76, 0})
	}()

	for i, char := range msg {
		if ctx.Err() != nil {
			return
		}
		switch char {
		case ' ':
			time.Sleep(interElementDuration * 7)
		default:
			ditDahs, ok := morse.IntlMorseCode.Encode(char)
			if !ok {
				fmt.Printf("Unsupported character: %c\n", char)
				continue
			}
			intervals := morse.DitDahsToIntervals(ditDahs, wpm)
			for _, interval := range intervals {
				if ctx.Err() != nil {
					return
				}
				if interval > 0 {
					midiOut.Send([]byte{0x90, 76, 127})
					time.Sleep(interval)
					midiOut.Send([]byte{0x80, 76, 0})
				} else {
					time.Sleep(-interval)
				}
			}
			if i < len(msg)-1 && msg[i+1] != ' ' {
				time.Sleep(interElementDuration * 3)
			}
		}
	}
	cancel()
}
