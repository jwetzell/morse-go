package main

import (
	"log/slog"
	"net"

	"github.com/jwetzell/morse-go"
	"github.com/jwetzell/morse-go/mopp"
)

func main() {

	addr, err := net.ResolveUDPAddr("udp", ":7373")
	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			slog.Error("Error reading from UDP", "err", err)
			continue
		}
		moppPacket := &mopp.Packet{}
		err = moppPacket.UnmarshalBinary(buf[:n])
		if err != nil {
			slog.Error("Error unmarshalling MOPP packet", "err", err)
			continue
		}
		receivedString, err := morse.IntlMorseCode.DecodeStates(moppPacket.States)
		if err != nil {
			slog.Error("Error decoding Morse code", "err", err)
			continue
		}
		slog.Info("Received string from", "serial", moppPacket.SerialNumber, "string", receivedString, "speed", moppPacket.Speed)
	}
}
