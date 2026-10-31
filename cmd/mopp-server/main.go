package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jwetzell/morse-go"
	"github.com/jwetzell/morse-go/mopp"
)

type ProtoType int

const (
	ProtoMopp ProtoType = iota
	ProtoText
)

type ChatClient struct {
	addr      net.Addr
	protoType ProtoType
	LastSeen  time.Time
}

type ChatServer struct {
	ctx      context.Context
	moppConn *net.UDPConn
	textConn *net.UDPConn
	clients  map[string]*ChatClient
}

func main() {

	moppAddr, err := net.ResolveUDPAddr("udp", ":7373")
	if err != nil {
		panic(err)
	}

	textAddr, err := net.ResolveUDPAddr("udp", ":7337")
	if err != nil {
		panic(err)
	}

	moppConn, err := net.ListenUDP("udp", moppAddr)
	if err != nil {
		panic(err)
	}
	defer moppConn.Close()

	textConn, err := net.ListenUDP("udp", textAddr)
	if err != nil {
		panic(err)
	}
	defer textConn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	chatServer := &ChatServer{
		ctx:      ctx,
		moppConn: moppConn,
		textConn: textConn,
		clients:  make(map[string]*ChatClient),
	}
	go chatServer.handleMoppClient()
	go chatServer.handleTextClient()
	//TODO(jwetzell): client keepalive?
	//TODO(jwetzell): dead client cleanup?
	<-ctx.Done()
}

func (cs *ChatServer) handleMoppClient() {
	ctx := cs.ctx
	moppConn := cs.moppConn
	buf := make([]byte, 1024)
	for ctx.Err() == nil {
		err := moppConn.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
		if err != nil {
			slog.Error("Error setting read deadline", "err", err)
			continue
		}
		n, rAddr, err := moppConn.ReadFromUDP(buf)
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
		_, ok := cs.clients[rAddr.String()]
		if !ok {
			cs.clients[rAddr.String()] = &ChatClient{addr: rAddr, protoType: ProtoMopp, LastSeen: time.Now()}
		}
		cs.handleInput(ctx, receivedString, rAddr, func(msg string, to net.Addr) {
			chatClient, ok := cs.clients[to.String()]
			if !ok {
				slog.Error("Client not found", "addr", to)
				return
			}

			var payload []byte
			var conn *net.UDPConn
			switch chatClient.protoType {
			case ProtoText:
				payload = []byte(msg + "\n")
				conn = cs.textConn
			case ProtoMopp:
				moppPacket := &mopp.Packet{
					ProtoVersion: 1,
					Speed:        20,
				}
				states, err := morse.IntlMorseCode.EncodeString(msg)
				if err != nil {
					slog.Error("Error encoding Morse code", "err", err)
					return
				}
				moppPacket.States = states
				data, err := moppPacket.MarshalBinary()
				if err != nil {
					slog.Error("Error marshalling MOPP packet", "err", err)
					return
				}
				payload = data
				conn = cs.moppConn
			default:
				slog.Error("Unknown protocol type", "proto", chatClient.protoType)
				return
			}
			slog.Info("Sending message to client", "addr", to, "proto", chatClient.protoType, "payload", payload)
			_, err = conn.WriteTo(payload, to)
			if err != nil {
				slog.Error("Error sending message to client", "err", err)
			}
		})
	}
}

func (cs *ChatServer) handleTextClient() {
	ctx := cs.ctx
	textConn := cs.textConn
	buf := make([]byte, 1024)
	for ctx.Err() == nil {
		err := textConn.SetReadDeadline(time.Now().Add(time.Millisecond * 200))
		if err != nil {
			slog.Error("Error setting read deadline", "err", err)
			continue
		}
		n, rAddr, err := textConn.ReadFromUDP(buf)
		if err != nil {
			if !os.IsTimeout(err) {
				slog.Error("Error reading from UDP", "err", err)
			}
			continue
		}
		trimmedString := strings.TrimSpace(string(buf[:n]))
		upperText := strings.ToUpper(trimmedString)

		for _, r := range upperText {
			switch r {
			case ' ':
				continue
			default:
				_, ok := morse.IntlMorseCode.EncodeRune(r)
				if !ok {
					slog.Error("Error encoding character", "char", string(r))
					continue
				}
			}
		}
		_, ok := cs.clients[rAddr.String()]
		if !ok {
			cs.clients[rAddr.String()] = &ChatClient{addr: rAddr, protoType: ProtoText}
		}
		cs.clients[rAddr.String()].LastSeen = time.Now()
		cs.handleInput(ctx, upperText, rAddr, func(msg string, to net.Addr) {

			chatClient, ok := cs.clients[to.String()]
			if !ok {
				slog.Error("Client not found", "addr", to)
				return
			}

			var payload []byte
			var conn *net.UDPConn
			switch chatClient.protoType {
			case ProtoText:
				payload = []byte(msg + "\n")
				conn = cs.textConn
			case ProtoMopp:
				moppPacket := &mopp.Packet{
					ProtoVersion: 1,
					Speed:        20,
				}
				states, err := morse.IntlMorseCode.EncodeString(msg)
				if err != nil {
					slog.Error("Error encoding Morse code", "err", err)
					return
				}
				moppPacket.States = states
				data, err := moppPacket.MarshalBinary()
				if err != nil {
					slog.Error("Error marshalling MOPP packet", "err", err)
					return
				}
				payload = data
				conn = cs.moppConn
			}
			_, err := conn.WriteTo(payload, to)
			if err != nil {
				slog.Error("Error sending message to client", "err", err)
			}
		})
	}
}

func (cs *ChatServer) handleInput(ctx context.Context, msg string, remoteAddr net.Addr, responseWriter func(msg string, to net.Addr)) {
	slog.Info("Received string", "from", remoteAddr, "string", msg)
	switch msg {
	// NOTE(jwetzell): handle special kinds of commands
	default:
		// Broadcast the message to all clients except the sender
		for addrStr, client := range cs.clients {
			if addrStr == remoteAddr.String() {
				continue
			}
			slog.Info("Sending message to client", "to", client.addr, "message", msg)
			responseWriter(msg, client.addr)
		}
	}

}
