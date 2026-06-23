package vband

import (
	"fmt"
	"strconv"
	"strings"
)

type PON struct{}

func (p *PON) Encode() ([]byte, error) {
	return []byte("PON"), nil
}

func (p *PON) Decode(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("invalid PON message")
	}
	if string(data[0:3]) != "PON" {
		return fmt.Errorf("invalid PON message")
	}
	return nil
}

type NOK struct{}

func (n *NOK) Encode() ([]byte, error) {
	return []byte("NOK"), nil
}

func (n *NOK) Decode(data []byte) error {
	if len(data) < 3 {
		return fmt.Errorf("invalid NOK message")
	}
	if string(data[0:3]) != "NOK" {
		return fmt.Errorf("invalid NOK message")
	}
	return nil
}

type DecodeMorse struct {
	Message string
}

func (d *DecodeMorse) Encode() ([]byte, error) {
	return []byte("DM," + d.Message), nil
}

func (d *DecodeMorse) Decode(data []byte) error {
	parts := strings.SplitN(string(data), ",", 2)
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

func (s *SpaceMark) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("SM,%d,%d", s.Space, s.Mark)), nil
}

func (s *SpaceMark) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
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

func (i *IncomingSpaceMark) Encode() ([]byte, error) {
	return []byte(fmt.Sprintf("SMK,%s,%s,%s,%d,%d", i.Channel, i.UserID, i.UserName, i.Space, i.Mark)), nil
}

func (i *IncomingSpaceMark) Decode(data []byte) error {
	parts := strings.Split(string(data), ",")
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
