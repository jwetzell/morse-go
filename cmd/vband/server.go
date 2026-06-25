package main

import (
	"log/slog"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/jwetzell/morse-go/vband"
)

type Server struct {
	userSockets map[*websocket.Conn]*User
	channels    []Channel
}

func handleConnect(conn *websocket.Conn, msg vband.Connect) {
	response := vband.ConnectOK{
		UserName:      msg.Name,
		UserID:        strconv.Itoa(int(latestUserID)),
		ServerVersion: "0.0.0",
	}
	user := User{
		ID:   strconv.Itoa(int(latestUserID)),
		Name: msg.Name,
	}
	userSockets[conn] = user
	latestUserID++
	responseString, err := response.Encode()
	if err != nil {
		slog.Error("Failed to encode ConnectOK message", "error", err)
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(responseString))
	if err != nil {
		slog.Error("Failed to send ConnectOK message", "error", err)
		return
	}
}
