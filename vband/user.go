package vband

import (
	"fmt"
	"strings"
)

type NameChange struct {
	UserName string
}

func (n *NameChange) Encode() ([]byte, error) {
	return []byte("NC," + n.UserName), nil
}

func (n *NameChange) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) != 2 || parts[0] != "NC" {
		return fmt.Errorf("invalid name change message")
	}

	n.UserName = parts[1]
	return nil
}
