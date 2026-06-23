package vband

import (
	"fmt"
	"strconv"
	"strings"
)

type PON struct{}

func (p *PON) Encode() (string, error) {
	return "PON", nil
}

func (p *PON) Decode(data string) error {
	if len(data) < 3 {
		return fmt.Errorf("invalid PON message")
	}
	if data[0:3] != "PON" {
		return fmt.Errorf("invalid PON message")
	}
	return nil
}

type NOK struct{}

func (n *NOK) Encode() (string, error) {
	return "NOK", nil
}

func (n *NOK) Decode(data string) error {
	if len(data) < 3 {
		return fmt.Errorf("invalid NOK message")
	}
	if data[0:3] != "NOK" {
		return fmt.Errorf("invalid NOK message")
	}
	return nil
}

type DecodeMorse struct {
	Message string
}

func (d *DecodeMorse) Encode() (string, error) {
	return "DM," + d.Message, nil
}

func (d *DecodeMorse) Decode(data string) error {
	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 || parts[0] != "DM" {
		return fmt.Errorf("invalid decode morse message")
	}

	d.Message = parts[1]
	return nil
}

type SpaceMark struct {
	Space uint
	Mark  uint
}

func (s *SpaceMark) Encode() (string, error) {
	return fmt.Sprintf("SM,%d,%d", s.Space, s.Mark), nil
}

func (s *SpaceMark) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 3 || parts[0] != "SM" {
		return fmt.Errorf("invalid space mark message")
	}

	space, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid space in space mark message: %w", err)
	}
	s.Space = uint(space)

	mark, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid mark in space mark message: %w", err)
	}
	s.Mark = uint(mark)

	return nil
}

type IncomingSpaceMark struct {
	Channel  string
	UserID   string
	UserName string
	Space    uint
	Mark     uint
}

func (i *IncomingSpaceMark) Encode() (string, error) {
	return fmt.Sprintf("SMK,%s,%s,%s,%d,%d", i.Channel, i.UserID, i.UserName, i.Space, i.Mark), nil
}

func (i *IncomingSpaceMark) Decode(data string) error {
	parts := strings.Split(data, ",")
	if len(parts) != 6 || parts[0] != "SMK" {
		return fmt.Errorf("invalid incoming space mark message")
	}

	i.Channel = parts[1]
	i.UserID = parts[2]
	i.UserName = parts[3]

	space, err := strconv.ParseUint(parts[4], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid space in incoming space mark message: %w", err)
	}
	i.Space = uint(space)

	mark, err := strconv.ParseUint(parts[5], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid mark in incoming space mark message: %w", err)
	}
	i.Mark = uint(mark)
	return nil
}
