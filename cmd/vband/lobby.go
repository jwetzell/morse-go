package main

import (
	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/jwetzell/morse-go/vband"
)

func handleLobbyUpdate(conn *websocket.Conn, msg vband.LobbyUpdate) {
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
			break
		}
	}
}

func handleLobbyConnect(conn *websocket.Conn) {
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
	}
	conn.WriteMessage(websocket.TextMessage, []byte("CLC"))
}
