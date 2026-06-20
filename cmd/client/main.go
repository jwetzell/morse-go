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

	"github.com/jwetzell/morse-go/kob"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

var host string
var port int
var wire int
var debug bool

var stationID string
var midiOut string

func init() {
	flag.StringVar(&host, "host", "mtc-kob.dyndns.org", "KOB server address")
	flag.IntVar(&port, "port", 7890, "KOB server port")
	flag.IntVar(&wire, "wire", 101, "Wire number to connect to")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.StringVar(&stationID, "station-id", "", "Station ID")
	flag.StringVar(&midiOut, "midi-out", "", "MIDI output device name (optional)")
}

func main() {
	defer midi.CloseDriver()
	flag.Parse()

	if stationID == "" {
		slog.Error("Station ID is required")
		return
	}

	var logLevel slog.Level = slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if host == "" {
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

	hostPort := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	udpAddr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		slog.Error("Failed to resolve server address", "error", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		slog.Error("Failed to connect to server", "error", err)
	}
	defer conn.Close()

	station := kob.NewStation(ctx, stationID, "kob-go 0.0.0")

	connectPacket := kob.ConnectPacket{Wire: uint16(wire)}

	idPacket := kob.IDPacket{
		StationID:  station.ID(),
		SequenceNo: 1,
		Flags:      0,
		Version:    station.Version(),
	}

	connectPacketTicker := time.NewTicker(time.Second * 10)

	data, err := connectPacket.MarshalBinary()
	if err != nil {
		slog.Error("Failed to marshal ConnectPacket", "error", err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		slog.Error("Failed to write ConnectPacket", "error", err)
		return
	}

	idData, err := idPacket.MarshalBinary()
	if err != nil {
		slog.Error("Failed to marshal IDPacket", "error", err)
		return
	}

	_, err = conn.Write(idData)
	if err != nil {
		slog.Error("Failed to write IDPacket", "error", err)
		return
	}

	go func() {
		defer func() {
			slog.Debug("keepalive packet sender stopped")
		}()
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
				connectPacketTicker.Stop()
				return
			case <-connectPacketTicker.C:
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
				idPacket.SequenceNo = idPacket.SequenceNo + 1
			default:
				continue
			}
		}
	}()
	kobWire := kob.NewWire(ctx)
	kobWire.Connect(station)

	var midiOutDevice drivers.Out

	if midiOut != "" {
		out, err := midi.FindOutPort(midiOut)
		if err != nil {
			fmt.Printf("can't find MIDI output device: %s\n", midiOut)
			return
		}

		err = out.Open()
		if err != nil {
			fmt.Printf("can't open MIDI output device: %s\n", midiOut)
			return
		}
		defer func() {
			midiOutDevice.Send([]byte{0x80, 76, 0})
			out.Close()
		}()

		midiOutDevice = out
	}

	go func() {
		defer func() {
			slog.Debug("wire watcher stopped")
		}()
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
				return
			case state := <-kobWire.State:
				var midiMsg []byte
				// TODO(jwetzell): make midi configurable
				if state {
					midiMsg = []byte{0x90, 76, 127}
					slog.Debug("wire state changed", "state", "open")
				} else {
					midiMsg = []byte{0x80, 76, 0}
					slog.Debug("wire state changed", "state", "closed")
				}
				if midiOutDevice != nil {
					err := midiOutDevice.Send(midiMsg)
					if err != nil {
						slog.Error("Failed to send MIDI message", "error", err)
					}
				}
			default:
				continue
			}
		}
	}()

	buffer := make([]byte, 2048)
	var lastSequenceNo uint32

	stations := make(map[string]*kob.Station)
	stations[station.ID()] = station

	defer func() {
		kobWire.Close()
	}()

	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down...")
			disconnectPacket := kob.DisconnectPacket{Wire: uint16(wire)}
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

			conn.SetDeadline(time.Now().Add(time.Millisecond * 100))

			n, err := conn.Read(buffer)
			if err != nil {
				continue
			}

			bytes := buffer[:n]

			if len(bytes) < 2 {
				slog.Error("Received data too short", "length", len(bytes))
			}

			commandCode := int(bytes[0]) | (int(bytes[1]) << 8)
			switch commandCode {
			case 0x2:
				var disconnectPacket kob.DisconnectPacket
				err := disconnectPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal DisconnectPacket", "error", err)
					continue
				}
				slog.Info("Received DisconnectPacket", "wire", disconnectPacket.Wire)
			case 0x3:
				var dataPacket kob.DataPacket
				err := dataPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal DataPacket", "error", err)
					continue
				}
				if dataPacket.SequenceNo == lastSequenceNo || dataPacket.SequenceNo < lastSequenceNo {
					continue
				}

				if len(dataPacket.CodeList) == 0 {
					var idPacket kob.IDPacket
					err := idPacket.UnmarshalBinary(bytes)
					if err != nil {
						slog.Error("Failed to unmarshal IDPacket", "error", err)
					}
					lastSequenceNo = idPacket.SequenceNo
					_, exists := stations[idPacket.StationID]
					if !exists {
						slog.Info("New station connected", "stationID", idPacket.StationID, "version", idPacket.Version)
						newStation := kob.NewStation(ctx, idPacket.StationID, idPacket.Version)
						stations[idPacket.StationID] = newStation
						kobWire.Connect(newStation)
					}
					continue
				} else {
					slog.Debug("Received DataPacket", "stationID", dataPacket.StationID, "sequenceNo", dataPacket.SequenceNo, "codeList", dataPacket.CodeList)
					lastSequenceNo = dataPacket.SequenceNo

					var stationForData *kob.Station
					existingStation, exists := stations[dataPacket.StationID]
					if !exists {
						slog.Info("Data seen from new station", "stationID", dataPacket.StationID)
						newStation := kob.NewStation(ctx, dataPacket.StationID, "")
						stations[dataPacket.StationID] = newStation
						kobWire.Connect(newStation)
						stationForData = newStation
					} else {
						stationForData = existingStation
					}

					if stationForData != nil {
						stationForData.PushCodeList(dataPacket.CodeList)
					}
				}
			case 0x4:
				var connectPacket kob.ConnectPacket
				err := connectPacket.UnmarshalBinary(bytes)
				if err != nil {
					slog.Error("Failed to unmarshal ConnectPacket", "error", err)
					continue
				}
				slog.Debug("Received ConnectPacket", "wire", connectPacket.Wire)
				continue
			case 0x5:
				var ackPacket kob.AckPacket
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
	<-ctx.Done()
}
