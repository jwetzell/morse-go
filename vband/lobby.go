package vband

import (
	"fmt"
	"strings"
)

type LobbyConnect struct{}

func (l *LobbyConnect) Encode() (string, error) {
	return "LC", nil
}

func (l *LobbyConnect) Decode(data string) error {
	if data != "LC" {
		return fmt.Errorf("invalid lobby connect message")
	}
	return nil
}

type LobbyUpdate struct {
	Channel string
}

func (l *LobbyUpdate) Encode() (string, error) {
	return fmt.Sprintf("LU,%s", l.Channel), nil
}

func (l *LobbyUpdate) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 2 || parts[0] != "LU" {
		return fmt.Errorf("invalid lobby update message")
	}

	l.Channel = parts[1]
	return nil
}
