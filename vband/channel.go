package vband

import (
	"fmt"
	"strconv"
	"strings"
)

type ChannelListBegin struct{}

func (c *ChannelListBegin) Encode() ([]byte, error) {
	return []byte("CLB"), nil
}

func (c *ChannelListBegin) Decode(data []byte) error {
	if string(data) != "CLB" {
		return fmt.Errorf("invalid channel list begin message")
	}
	return nil
}

type ChannelListComplete struct{}

func (c *ChannelListComplete) Encode() ([]byte, error) {
	return []byte("CLC"), nil
}

func (c *ChannelListComplete) Decode(data []byte) error {
	if string(data) != "CLC" {
		return fmt.Errorf("invalid channel list complete message")
	}
	return nil
}

type ChannelListEntry struct {
	Name  string
	Count uint
}

func (c *ChannelListEntry) Encode() ([]byte, error) {
	if c.Count == 0 {
		return []byte(fmt.Sprintf("CLE,%s", c.Name)), nil
	}
	return []byte(fmt.Sprintf("CLE,%s,%d", c.Name, c.Count)), nil
}

func (c *ChannelListEntry) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) < 2 || parts[0] != "CLE" {
		return fmt.Errorf("invalid channel list entry message")
	}

	c.Name = parts[1]
	if len(parts) == 3 {
		var err error
		channelCount, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid count in channel list entry message: %w", err)
		}
		c.Count = uint(channelCount)
	}

	return nil
}

type ChannelListUpdate struct {
	Name  string
	Count uint
}

func (c *ChannelListUpdate) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("CLU,%s,%d", c.Name, c.Count)), nil
}

func (c *ChannelListUpdate) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) != 3 || parts[0] != "CLU" {
		return fmt.Errorf("invalid channel list update message")
	}

	c.Name = parts[1]
	channelCount, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid count in channel list update message: %w", err)
	}
	c.Count = uint(channelCount)

	return nil
}

type ChannelJoin struct {
	Name       string
	Attributes string
}

func (c *ChannelJoin) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("CJN,%s,%s", c.Name, c.Attributes)), nil
}

func (c *ChannelJoin) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) != 3 || parts[0] != "CJN" {
		return fmt.Errorf("invalid channel join message")
	}

	c.Name = parts[1]
	c.Attributes = parts[2]

	return nil
}

type ChannelJoinRequest struct {
	Name string
}

func (c *ChannelJoinRequest) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("JC,%s", c.Name)), nil
}

func (c *ChannelJoinRequest) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) != 2 || parts[0] != "JC" {
		return fmt.Errorf("invalid channel join request message")
	}

	c.Name = parts[1]

	return nil
}

type ListUpdate struct {
	ChannelName string
}

func (l *ListUpdate) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("LU,%s", l.ChannelName)), nil
}

func (l *ListUpdate) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
	if len(parts) != 2 || parts[0] != "LU" {
		return fmt.Errorf("invalid list update message")
	}

	l.ChannelName = parts[1]

	return nil
}
