package main

import (
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/jwetzell/morse-go/vband"
)

func handleJoinChannel(conn *websocket.Conn, msg vband.ChannelJoinRequest) {
	user, ok := userSockets[conn]
	if !ok {
		slog.Error("User not found for connection")
		return
	}

	for i, channel := range channels {
		if channel.Name == msg.Name {
			if channel.HasUser(user.ID) {
				slog.Warn("User already in channel", "user", user, "channel", channel.Name)
				continue
			}
			channels[i].AddUser(user)
			slog.Info("User added to channel", "user", user, "channel", channel.Name)

			channelJoin := vband.ChannelJoin{
				Name:       channel.Name,
				Attributes: "N",
			}
			responseString, err := channelJoin.Encode()
			if err != nil {
				slog.Error("Failed to encode ChannelJoin message", "error", err)
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
			if err != nil {
				slog.Error("Failed to send ChannelJoin message", "error", err)
				continue
			}
		} else {
			if channel.HasUser(user.ID) {
				channels[i].RemoveUser(user.ID)
			}
		}
	}
}
