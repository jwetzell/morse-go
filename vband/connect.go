package vband

import (
	"fmt"
	"strings"
)

type Connect struct {
	Name          string
	DeviceID      string
	ClientVersion string
}

func (c *Connect) Encode() (string, error) {
	return "CN," + c.Name + "," + c.DeviceID + "," + c.ClientVersion, nil
}

func (c *Connect) Decode(data string) error {
	parts := strings.Split(data, ",")

	if len(parts) != 4 || parts[0] != "CN" {
		return fmt.Errorf("invalid connect message")
	}

	c.Name = parts[1]
	c.DeviceID = parts[2]
	c.ClientVersion = parts[3]

	return nil
}

type ConnectOK struct {
	UserID        string
	UserName      string
	ServerVersion string // maybe?
}

func (c *ConnectOK) Encode() (string, error) {
	return "COK," + c.UserID + "," + c.UserName + "," + c.ServerVersion, nil
}

func (c *ConnectOK) Decode(data string) error {
	parts := strings.Split(data, ",")

	if len(parts) != 4 || parts[0] != "COK" {
		return fmt.Errorf("invalid connect OK message")
	}

	c.UserID = parts[1]
	c.UserName = parts[2]
	c.ServerVersion = parts[3]

	return nil
}
