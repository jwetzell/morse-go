package main

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jwetzell/morse-go/vband"
)

var ip string
var port uint

var latestUserID uint = 1000

var userSockets = make(map[*websocket.Conn]User)

var upgrader = websocket.Upgrader{
	Subprotocols: []string{"lws-hrs-vband2", "lws-hrs-vband2-ssl"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {

	flag.StringVar(&ip, "ip", "localhost", "IP address to listen on")
	flag.UintVar(&port, "port", 7385, "Port to listen on")
}

func main() {
	flag.Parse()

	if port < 1 || port > 65535 {
		panic("Port must be between 1 and 65535")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Received connection from", "remoteAddr", r.RemoteAddr)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Failed to upgrade connection", "error", err)
			return
		}
		defer conn.Close()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("Failed to read message", "error", err)
				break
			}

			slog.Info("Received message", "remoteAddr", r.RemoteAddr, "messageType", messageType, "message", string(message))
			if messageType == websocket.TextMessage {
				msgString := string(message)
				stringParts := strings.Split(msgString, ",")
				switch stringParts[0] {
				case "CN":
					var msg vband.Connect
					err := msg.Decode(msgString)
					if err != nil {
						slog.Error("Failed to decode Connect message", "error", err)
						continue
					}
					handleConnect(conn, msg)
				case "LC":
					handleLobbyConnect(conn)
				case "LU":
					var msg vband.LobbyUpdate
					err := msg.Decode(msgString)
					if err != nil {
						slog.Error("Failed to decode LobbyUpdate message", "error", err)
						continue
					}
					handleLobbyUpdate(conn, msg)
				case "JC":
					var msg vband.ChannelJoinRequest
					err := msg.Decode(msgString)
					if err != nil {
						slog.Error("Failed to decode ChannelJoinRequest message", "error", err)
						continue
					}
					handleJoinChannel(conn, msg)
				default:
					slog.Warn("Unknown message type", "message", msgString)
				}
			}
		}
		slog.Info("Connection closed", "remoteAddr", r.RemoteAddr)
	})
	http.ListenAndServe(net.JoinHostPort(ip, strconv.Itoa(int(port))), nil)
}
