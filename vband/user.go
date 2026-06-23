package vband

import (
	"fmt"
	"strings"
)

type NameChange struct {
	UserName string
}

func (n *NameChange) Encode() (string, error) {
	return "NC," + n.UserName, nil
}

func (n *NameChange) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 2 || parts[0] != "NC" {
		return fmt.Errorf("invalid name change message")
	}

	n.UserName = parts[1]
	return nil
}
