package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jwetzell/morse-go"
	"github.com/jwetzell/morse-go/mopp"
)

var hostFlag string
var portFlag int
var speedFlag int

func init() {
	flag.StringVar(&hostFlag, "host", "", "the host to send MOPP packets to")
	flag.IntVar(&portFlag, "port", 7373, "the port to send MOPP packets to")
	flag.IntVar(&speedFlag, "speed", 20, "the WPM to send MOPP packets at")
}

func main() {
	flag.Parse()
	if hostFlag == "" {
		fmt.Println("Error: host is required")
		return
	}
	if portFlag <= 0 || portFlag > 65535 {
		fmt.Println("Error: invalid port")
		return
	}
	if speedFlag <= 0 || speedFlag > mopp.MaxSpeed {
		fmt.Println("Error: invalid speed")
		return
	}

	client, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(hostFlag),
		Port: portFlag,
	})
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.Close()
	go clientInput(context.Background(), client)
	scanner := bufio.NewScanner(os.Stdin)
	serialNum := uint8(1)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Process the input line here
		words := strings.Split(line, " ")
		for _, word := range words {
			moppPacket := &mopp.Packet{
				ProtoVersion: 1,
				SerialNumber: serialNum,
				Speed:        uint8(speedFlag),
			}
			serialNum++
			if serialNum > mopp.MaxSerialNumber {
				serialNum = 1
			}
			states, err := morse.IntlMorseCode.EncodeString(strings.ToUpper(word))
			if err != nil {
				fmt.Println("Error encoding line to Morse code:", err)
				continue
			}
			moppPacket.States = states
			packetBytes, err := moppPacket.MarshalBinary()
			if err != nil {
				fmt.Println("Error marshaling MOPP packet:", err)
				continue
			}
			_, err = client.Write(packetBytes)
			if err != nil {
				fmt.Println("Error sending data to server:", err)
				continue
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}

func clientInput(ctx context.Context, client *net.UDPConn) {
	buf := make([]byte, 1024)
	for ctx.Err() == nil {
		err := client.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
		if err != nil {
			slog.Error("Error setting read deadline", "err", err)
			continue
		}
		n, _, err := client.ReadFromUDP(buf)
		if err != nil {
			if !os.IsTimeout(err) {
				slog.Error("Error reading from UDP", "err", err)
			}
			continue
		}
		moppPacket := &mopp.Packet{}
		err = moppPacket.UnmarshalBinary(buf[:n])
		if err != nil {
			slog.Error("Error unmarshalling MOPP packet", "err", err)
			continue
		}
		receivedString, err := morse.IntlMorseCode.DecodeString(moppPacket.States)
		if err != nil {
			slog.Error("Error decoding Morse code", "err", err)
			continue
		}
		os.Stdout.WriteString(receivedString)
	}
}
