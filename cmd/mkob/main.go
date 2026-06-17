package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jwetzell/morse-go"
	"github.com/jwetzell/morse-go/mkob"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var server string
var port int
var wire int
var debug bool

var ditMax int
var wordSpace int

var midiOutPort string

func init() {
	flag.StringVar(&server, "server", "mtc-kob.dyndns.org", "MTC-KOB server address")
	flag.IntVar(&port, "port", 7890, "MTC-KOB server port")
	flag.IntVar(&wire, "wire", 101, "Wire number to connect to")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")

	flag.IntVar(&ditMax, "dit-max", 100, "Maximum code list value to consider as a dit (default: 100)")
	flag.IntVar(&wordSpace, "word-space", 400, "Minimum code list value to consider as a word space (default: 400)")

	flag.StringVar(&midiOutPort, "midi-out", "", "MIDI output port name (optional)")
}

var sendFunc func([]byte) error = func(data []byte) error {
	return nil
}

func main() {
	flag.Parse()

	var logLevel slog.Level = slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if server == "" {
		slog.Error("Server address is required")
		return
	}

	if port <= 0 || port > 65535 {
		slog.Error("Invalid port number")
		return
	}

	if wire <= 0 || wire > 65535 {
		slog.Error("Invalid wire number")
		return
	}

	if midiOutPort != "" {
		out, err := midi.FindOutPort(midiOutPort)
		if err != nil {
			slog.Error("can't find midi output", "port", midiOutPort, "error", err)
			return
		}

		err = out.Open()
		if err != nil {
			slog.Error("can't open midi output", "port", midiOutPort, "error", err)
			return
		}

		sendFunc = out.Send
	}

	conn, err := net.Dial("udp", net.JoinHostPort(server, fmt.Sprintf("%d", port)))
	if err != nil {
		slog.Error("Failed to connect to server", "error", err)
	}
	defer conn.Close()

	connectPacket := mkob.ConnectPacket{Wire: uint16(wire)}

	idPacket := mkob.IDPacket{
		StationID:  "KD9PUI",
		SequenceNo: 1,
		Flags:      0,
		Version:    "mkob-go 0.0.0",
	}

	connectPacketTicker := time.NewTicker(time.Second * 5)

	data, err := connectPacket.MarshalBinary()
	if err != nil {
		slog.Error("Failed to marshal ConnectPacket", "error", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		slog.Error("Failed to write ConnectPacket", "error", err)
	}

	idData, err := idPacket.MarshalBinary()
	if err != nil {
		slog.Error("Failed to marshal IDPacket", "error", err)

	}

	_, err = conn.Write(idData)
	if err != nil {
		slog.Error("Failed to write IDPacket", "error", err)

	}

	go func() {
		for range connectPacketTicker.C {
			data, err := connectPacket.MarshalBinary()
			if err != nil {
				slog.Error("Failed to marshal ConnectPacket", "error", err)
				continue
			}

			_, err = conn.Write(data)
			if err != nil {
				slog.Error("Failed to write ConnectPacket", "error", err)
				continue
			}

			idData, err := idPacket.MarshalBinary()
			if err != nil {
				slog.Error("Failed to marshal IDPacket", "error", err)
				continue
			}

			_, err = conn.Write(idData)
			if err != nil {
				slog.Error("Failed to write IDPacket", "error", err)
				continue
			}
		}
	}()

	mkobWire := mkob.NewWire()

	go func() {
		for state := range mkobWire.State {

			var msg []byte
			if midiOutPort != "" {
				// TODO(jwetzell): make this configurable
				if state {
					msg = []byte{0x90, 76, 127}
				} else {
					msg = []byte{0x80, 76, 0}
				}
			}
			if msg != nil {
				err := sendFunc(msg)
				if err != nil {
					slog.Error("Failed to send event", "error", err)
				}
			}
		}
	}()

	buffer := make([]byte, 2048)
	var lastSequenceNo uint32
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down...")
			disconnectPacket := mkob.DisconnectPacket{Wire: uint16(wire)}
			data, err := disconnectPacket.MarshalBinary()
			if err != nil {
				slog.Error("Failed to marshal DisconnectPacket", "error", err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				slog.Error("Failed to write DisconnectPacket", "error", err)
			}
			return
		default:

			conn.SetDeadline(time.Now().Add(time.Second * 100))

			n, err := conn.Read(buffer)
			if err != nil {
				slog.Error("Failed to read", "error", err)
			}

			bytes := buffer[:n]

			if len(bytes) < 2 {
				slog.Error("Received data too short", "length", len(bytes))
			}

			commandCode := int(bytes[0]) | (int(bytes[1]) << 8)
			switch commandCode {
			case 0x2:
				var disconnectPacket mkob.DisconnectPacket
				err := disconnectPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal DisconnectPacket", "error", err)
					continue
				}
				slog.Info("Received DisconnectPacket", "wire", disconnectPacket.Wire)
				return
			case 0x3:
				var dataPacket mkob.DataPacket
				err := dataPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal DataPacket", "error", err)
					continue
				}
				if dataPacket.SequenceNo == lastSequenceNo || dataPacket.SequenceNo < lastSequenceNo {
					continue
				}

				if len(dataPacket.CodeList) == 0 {
					var idPacket mkob.IDPacket
					err := idPacket.UnmarshalBinary(bytes)
					if err != nil {
						slog.Error("Failed to unmarshal IDPacket", "error", err)
					}
					lastSequenceNo = idPacket.SequenceNo
					continue
				} else {
					ditDahs := morse.CodeListToDitDahs(dataPacket.CodeList, int32(ditMax))

					letter, err := morse.ASCIIFromDitDahs(ditDahs)
					if err != nil {
						slog.Error("Failed to convert dit-dahs to ASCII", "error", err)
						continue
					}
					if dataPacket.CodeList[0] < -int32(wordSpace) {
						os.Stdout.Write([]byte(" "))
					}
					slog.Debug("Received DataPacket", "stationID", dataPacket.StationID, "sequenceNo", dataPacket.SequenceNo, "codeList", dataPacket.CodeList, "ditDahs", ditDahs, "letter", letter)
					os.Stdout.Write([]byte(letter))
					mkobWire.RegisterCodeList(dataPacket.CodeList)
					lastSequenceNo = dataPacket.SequenceNo
				}
			case 0x4:
				var connectPacket mkob.ConnectPacket
				err := connectPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal ConnectPacket", "error", err)
					continue
				}
				slog.Debug("Received ConnectPacket", "wire", connectPacket.Wire)
				continue
			case 0x5:
				var ackPacket mkob.AckPacket
				err := ackPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal AckPacket", "error", err)
					continue
				}
				slog.Debug("Received AckPacket")
			default:
				slog.Error("Unknown command code", "code", commandCode)
			}
		}
	}
}
