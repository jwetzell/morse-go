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
					slog.Info("Decoded Connect message", "message", msg)
					response := vband.ConnectOK{
						UserName:      msg.Name,
						UserID:        strconv.Itoa(int(latestUserID)),
						ServerVersion: "0.0.0",
					}
					latestUserID++
					responseString, err := response.Encode()
					if err != nil {
						slog.Error("Failed to encode ConnectOK message", "error", err)
						continue
					}
					err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
					if err != nil {
						slog.Error("Failed to send ConnectOK message", "error", err)
						continue
					}
					slog.Info("Sent ConnectOK message", "message", responseString)
				case "LC":
					conn.WriteMessage(websocket.TextMessage, []byte("CLB"))
					for _, channel := range channels {
						channelListEntry := vband.ChannelListEntry{
							Name:  channel.Name,
							Count: uint(len(channel.Users)),
						}
						responseString, err := channelListEntry.Encode()
						if err != nil {
							slog.Error("Failed to encode ChannelListEntry message", "error", err)
							continue
						}
						err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
						if err != nil {
							slog.Error("Failed to send ChannelListEntry message", "error", err)
							continue
						}
						slog.Info("Sent ChannelListEntry message", "message", responseString)
					}
					conn.WriteMessage(websocket.TextMessage, []byte("CLC"))
				case "LU":
					var msg vband.LobbyUpdate
					err := msg.Decode(msgString)
					if err != nil {
						slog.Error("Failed to decode LobbyUpdate message", "error", err)
						continue
					}
					for _, channel := range channels {
						if channel.Name == msg.Channel {
							userListBegin := vband.UserListBegin{
								Channel: channel.Name,
							}
							responseString, err := userListBegin.Encode()
							if err != nil {
								slog.Error("Failed to encode UserListBegin message", "error", err)
								continue
							}
							err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
							if err != nil {
								slog.Error("Failed to send UserListBegin message", "error", err)
								continue
							}
							slog.Info("Sent UserListBegin message", "message", responseString)

							for _, user := range channel.Users {
								userListEntry := vband.UserListEntry{
									Channel:  channel.Name,
									UserID:   user.ID,
									UserName: user.Name,
								}
								responseString, err := userListEntry.Encode()
								if err != nil {
									slog.Error("Failed to encode UserListEntry message", "error", err)
									break
								}
								err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
								if err != nil {
									slog.Error("Failed to send UserListEntry message", "error", err)
									break
								}
								slog.Info("Sent UserListEntry message", "message", responseString)
							}

							userListComplete := vband.UserListComplete{
								Channel: channel.Name,
							}
							responseString, err = userListComplete.Encode()
							if err != nil {
								slog.Error("Failed to encode UserListComplete message", "error", err)
								break
							}
							err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
							if err != nil {
								slog.Error("Failed to send UserListComplete message", "error", err)
								break
							}
							slog.Info("Sent UserListComplete message", "message", responseString)
							break
						}
					}
				case "JC":
					var msg vband.ChannelJoinRequest
					err := msg.Decode(msgString)
					if err != nil {
						slog.Error("Failed to decode ChannelJoinRequest message", "error", err)
						continue
					}
					slog.Info("Decoded ChannelJoinRequest message", "message", msg)
				default:
					slog.Warn("Unknown message type", "message", msgString)
				}
			}
		}
		slog.Info("Connection closed", "remoteAddr", r.RemoteAddr)
	})

	http.ListenAndServe(net.JoinHostPort(ip, strconv.Itoa(int(port))), nil)
}
