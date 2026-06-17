package kob

import (
	"encoding"
	"fmt"
	"strings"
	"unicode"
)

type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type ConnectPacket struct {
	Wire uint16
}

func (cp *ConnectPacket) MarshalBinary() ([]byte, error) {
	return []byte{0x4, 0x0, byte(cp.Wire & 0xff), byte(cp.Wire >> 8)}, nil
}

func (cp *ConnectPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("data too short for ConnectPacket")
	}
	cp.Wire = uint16(data[2]) | uint16(data[3])<<8
	return nil
}

type DisconnectPacket struct {
	Wire uint16
}

func (cp *DisconnectPacket) MarshalBinary() ([]byte, error) {
	return []byte{0x4, 0x0, byte(cp.Wire & 0xff), byte(cp.Wire >> 8)}, nil
}

func (cp *DisconnectPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return fmt.Errorf("data too short for DisconnectPacket")
	}
	cp.Wire = uint16(data[2]) | uint16(data[3])<<8
	return nil
}

type IDPacket struct {
	StationID  string
	SequenceNo uint32
	Flags      uint32
	Version    string
}

func (ip *IDPacket) MarshalBinary() ([]byte, error) {

	if !isASCII(ip.StationID) {
		return nil, fmt.Errorf("station ID must be ASCII")
	}

	if !isASCII(ip.Version) {
		return nil, fmt.Errorf("version must be ASCII")
	}

	if len(ip.StationID) > 127 {
		return nil, fmt.Errorf("station ID too long (max 127 bytes)")
	}

	if len(ip.Version) > 127 {
		return nil, fmt.Errorf("version too long (max 127 bytes)")
	}

	packetLength := 2 + 2 + 128 + 4 + 4 + 4 + 216 + 128 + 8

	bytes := make([]byte, packetLength)

	// Command code
	bytes[0] = 0x3
	bytes[1] = 0
	// byte count
	bytes[2] = byte((packetLength - 4) & 0xff)
	bytes[3] = byte((packetLength - 4) >> 8)

	// Station ID (128 bytes, null-padded) - max 127 chars + null terminator
	copy(bytes[4:132], []byte(ip.StationID))

	// Sequence number
	bytes[136] = byte(ip.SequenceNo & 0xff)
	bytes[137] = byte((ip.SequenceNo >> 8) & 0xff)
	bytes[138] = byte((ip.SequenceNo >> 16) & 0xff)
	bytes[139] = byte((ip.SequenceNo >> 24) & 0xff)

	// Flags
	bytes[140] = byte(ip.Flags & 0xff)
	bytes[141] = byte((ip.Flags >> 8) & 0xff)
	bytes[142] = byte((ip.Flags >> 16) & 0xff)
	bytes[143] = byte((ip.Flags >> 24) & 0xff)

	// Version (128 bytes, null-padded) - max 127 chars + null terminator
	copy(bytes[360:488], []byte(ip.Version))

	return bytes, nil
}

func (ip *IDPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 496 {
		return fmt.Errorf("data too short for IDPacket")
	}

	if data[0] != 0x3 || data[1] != 0 {
		return fmt.Errorf("invalid command code for IDPacket")
	}

	ip.StationID = string(data[4:132])
	nullByteIndex := strings.IndexByte(ip.StationID, 0)
	if nullByteIndex != -1 {
		ip.StationID = ip.StationID[:nullByteIndex]
	}
	ip.SequenceNo = uint32(data[136]) | uint32(data[137])<<8 | uint32(data[138])<<16 | uint32(data[139])<<24
	ip.Flags = uint32(data[140]) | uint32(data[141])<<8 | uint32(data[142])<<16 | uint32(data[143])<<24
	ip.Version = string(data[360:488])
	nullByteIndex = strings.IndexByte(ip.Version, 0)
	if nullByteIndex != -1 {
		ip.Version = ip.Version[:nullByteIndex]
	}
	return nil
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

type DataPacket struct {
	StationID  string
	SequenceNo uint32
	CodeList   []int32
}

func (dp *DataPacket) MarshalBinary() ([]byte, error) {

	if !isASCII(dp.StationID) {
		return nil, fmt.Errorf("station ID must be ASCII")
	}

	if len(dp.StationID) > 127 {
		return nil, fmt.Errorf("station ID too long (max 127 bytes)")
	}

	packetLength := 2 + 2 + 128 + 4 + 4 + 4 + 216 + 128 + 8

	bytes := make([]byte, packetLength)

	// Command code
	bytes[0] = 0x3
	bytes[1] = 0
	// byte count
	bytes[2] = byte((packetLength - 4) & 0xff)
	bytes[3] = byte((packetLength - 4) >> 8)

	// Station ID (128 bytes, null-padded) - max 127 chars + null terminator
	copy(bytes[4:132], []byte(dp.StationID))

	// Sequence number
	bytes[136] = byte(dp.SequenceNo & 0xff)
	bytes[137] = byte((dp.SequenceNo >> 8) & 0xff)
	bytes[138] = byte((dp.SequenceNo >> 16) & 0xff)
	bytes[139] = byte((dp.SequenceNo >> 24) & 0xff)

	codeListCount := len(dp.CodeList)
	if codeListCount > 51 {
		return nil, fmt.Errorf("code list too long (max 51 entries)")
	}

	// CodeList (216 bytes, up to 51 uint32 entries)
	for i, code := range dp.CodeList {
		offset := 152 + i*4
		bytes[offset] = byte(code & 0xff)
		bytes[offset+1] = byte((code >> 8) & 0xff)
		bytes[offset+2] = byte((code >> 16) & 0xff)
		bytes[offset+3] = byte((code >> 24) & 0xff)
	}

	bytes[356] = byte(codeListCount & 0xff)
	bytes[357] = byte((codeListCount >> 8) & 0xff)
	bytes[358] = byte((codeListCount >> 16) & 0xff)
	bytes[359] = byte((codeListCount >> 24) & 0xff)

	return bytes, nil
}

func (dp *DataPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 496 {
		return fmt.Errorf("data too short for DataPacket")
	}

	if data[0] != 0x3 || data[1] != 0 {
		return fmt.Errorf("invalid command code for DataPacket")
	}

	dp.StationID = string(data[4:132])

	nullByteIndex := strings.IndexByte(dp.StationID, 0)
	if nullByteIndex != -1 {
		dp.StationID = dp.StationID[:nullByteIndex]
	}
	dp.SequenceNo = uint32(data[136]) | uint32(data[137])<<8 | uint32(data[138])<<16 | uint32(data[139])<<24
	codeListCount := uint32(data[356]) | uint32(data[357])<<8 | uint32(data[358])<<16 | uint32(data[359])<<24

	if codeListCount > 51 {
		return fmt.Errorf("code list count too large: %d (max 51)", codeListCount)
	}
	dp.CodeList = make([]int32, codeListCount)
	for i := 0; i < int(codeListCount); i++ {
		offset := 152 + i*4
		dp.CodeList[i] = int32(data[offset]) | int32(data[offset+1])<<8 | int32(data[offset+2])<<16 | int32(data[offset+3])<<24
	}
	return nil
}

type AckPacket struct{}

func (ap *AckPacket) MarshalBinary() ([]byte, error) {
	return []byte{0x5, 0x0}, nil
}

func (ap *AckPacket) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("data too short for AckPacket")
	}
	if data[0] != 0x5 || data[1] != 0 {
		return fmt.Errorf("invalid command code for AckPacket")
	}
	return nil
}
