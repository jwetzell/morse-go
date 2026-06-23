package vband

import (
	"fmt"
	"strings"
)

type UserListBegin struct {
	Channel string
}

func (u *UserListBegin) Encode() (string, error) {
	return "ULB," + u.Channel, nil
}

func (u *UserListBegin) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 2 || parts[0] != "ULB" {
		return fmt.Errorf("invalid user list begin message")
	}

	u.Channel = parts[1]
	return nil
}

type UserListEntry struct {
	Channel  string
	UserID   string
	UserName string
}

func (u *UserListEntry) Encode() (string, error) {
	return "ULE," + u.Channel + "," + u.UserID + "," + u.UserName, nil
}

func (u *UserListEntry) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 4 || parts[0] != "ULE" {
		return fmt.Errorf("invalid user list entry message")
	}

	u.Channel = parts[1]
	u.UserID = parts[2]
	u.UserName = parts[3]

	return nil
}

type UserListComplete struct {
	Channel string
}

func (u *UserListComplete) Encode() (string, error) {
	return "ULC," + u.Channel, nil
}

func (u *UserListComplete) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 2 || parts[0] != "ULC" {
		return fmt.Errorf("invalid user list complete message")
	}

	u.Channel = parts[1]
	return nil
}
