// Morse Over Packet Protocol - https://github.com/oe1wkl/Morserino-32/blob/master/Documentation/Protocol%20Description/morse_code_over_packet_protocol.md
package mopp

import (
	"fmt"

	"github.com/jwetzell/morse-go"
)

const MaxSerialNumber = 0b111111
const MaxSpeed = 0b111111

type Packet struct {
	ProtoVersion uint8
	SerialNumber uint8
	Speed        uint8
	States       []morse.State
}

func (p *Packet) UnmarshalBinary(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("data too short")
	}

	p.ProtoVersion = (data[0] & 0b11000000) >> 6
	if p.ProtoVersion != 0b01 {
		return fmt.Errorf("unsupported protocol version: %b", p.ProtoVersion)
	}
	p.SerialNumber = (data[0] & 0b00111111)
	p.Speed = (data[1] & 0b11111100) >> 2
	remnantState := (data[1] & 0b00000011)
	morseState, err := moppStateToMorse(remnantState)
	if err != nil {
		return err
	}
	p.States = []morse.State{morseState}
	for i := 2; i < len(data); i++ {
		for stateIndex := 0; stateIndex < 4; stateIndex++ {
			shift := 6 - 2*stateIndex
			moppState := (data[i] & uint8(0b11<<shift)) >> shift
			morseState, err := moppStateToMorse(moppState)
			if err != nil {
				return err
			}
			p.States = append(p.States, morseState)
			if morseState == morse.StateEOW {
				return nil
			}
		}
	}
	return nil
}

func (p *Packet) MarshalBinary() ([]byte, error) {
	if p.ProtoVersion != 0b01 {
		return nil, fmt.Errorf("unsupported protocol version: %b", p.ProtoVersion)
	}

	if p.Speed == 0 || p.Speed > MaxSpeed {
		return nil, fmt.Errorf("unsupported speed: %b", p.Speed)
	}

	if p.SerialNumber > MaxSerialNumber {
		return nil, fmt.Errorf("unsupported serial number: %b", p.SerialNumber)
	}

	data := []byte{0, 0}
	data[0] = (p.ProtoVersion << 6) | (p.SerialNumber & 0b00111111)
	data[1] = (p.Speed << 2)

	if len(p.States) == 0 {
		data[1] |= byte(morse.StateEOW)
		return data, nil
	}

	remnantState, err := morseStateToMopp(p.States[0])
	if err != nil {
		return nil, err
	}
	data[1] |= remnantState

	stateIndex := 1
	for i := 2; stateIndex < len(p.States); i++ {
		data = append(data, 0)
		for slot := 0; slot < 4 && stateIndex < len(p.States); slot++ {
			shift := 6 - 2*slot
			moppState, err := morseStateToMopp(p.States[stateIndex])
			if err != nil {
				return nil, err
			}
			data[i] |= moppState << shift
			//NOTE(jwetzell): should a MOPP packet be able to contain multiple "words"? is extra states an errorable thing?
			if p.States[stateIndex] == morse.StateEOW {
				return data, nil
			}
			stateIndex++
		}
	}

	return data, nil
}

func morseStateToMopp(state morse.State) (uint8, error) {
	switch state {
	case morse.StateDit:
		return 0b01, nil
	case morse.StateDah:
		return 0b10, nil
	case morse.StateEOC:
		return 0b00, nil
	case morse.StateEOW:
		return 0b11, nil
	default:
		return 0, fmt.Errorf("invalid Morse state: %v", state)
	}
}

func moppStateToMorse(state uint8) (morse.State, error) {
	switch state {
	case 0b01:
		return morse.StateDit, nil
	case 0b10:
		return morse.StateDah, nil
	case 0b00:
		return morse.StateEOC, nil
	case 0b11:
		return morse.StateEOW, nil
	default:
		return 0, fmt.Errorf("invalid MOPP state: %b", state)
	}
}
