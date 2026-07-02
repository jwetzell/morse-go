package vband

import (
	"fmt"
	"strconv"
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
	UserID        uint
	UserName      string
	ServerVersion string // maybe?
}

func (c *ConnectOK) Encode() (string, error) {
	return fmt.Sprintf("COK,%d,%s,%s", c.UserID, c.UserName, c.ServerVersion), nil
}

func (c *ConnectOK) Decode(data string) error {
	parts := strings.Split(data, ",")

	if len(parts) != 4 || parts[0] != "COK" {
		return fmt.Errorf("invalid connect OK message")
	}

	userID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}
	c.UserID = uint(userID)
	c.UserName = parts[2]
	c.ServerVersion = parts[3]

	return nil
}
