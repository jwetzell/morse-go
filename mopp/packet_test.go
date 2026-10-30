package mopp_test

import (
	"testing"

	"github.com/jwetzell/morse-go"
	"github.com/jwetzell/morse-go/mopp"
	"github.com/stretchr/testify/require"
)

func TestGoodPacketUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name   string
		bytes  []byte
		packet mopp.Packet
	}{
		{
			name:  "PARIS",
			bytes: []byte{0x5b, 0x41, 0xa4, 0x61, 0x91, 0x45, 0x70},
			packet: mopp.Packet{
				ProtoVersion: 0b01,
				SerialNumber: 0b011011,
				Speed:        0b010000,
				States: []morse.State{
					morse.StateDit,
					morse.StateDah,
					morse.StateDah,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDah,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDah,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDit,
					morse.StateDit,
					morse.StateEOW,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mopPacket := &mopp.Packet{}

			err := mopPacket.UnmarshalBinary(tc.bytes)
			require.NoError(t, err)
			require.Equal(t, &tc.packet, mopPacket)
		})
	}
}

func TestBadPacketUnmarshalBinary(t *testing.T) {
	testCases := []struct {
		name        string
		bytes       []byte
		errorString string
	}{
		{
			name:        "not enough bytes",
			bytes:       []byte{0x5b},
			errorString: "data too short",
		},
		{
			name:        "unsupported protocol version",
			bytes:       []byte{0x00, 0x00},
			errorString: "unsupported protocol version: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mopPacket := &mopp.Packet{}

			err := mopPacket.UnmarshalBinary(tc.bytes)
			require.Error(t, err)
			require.EqualError(t, err, tc.errorString)
		})
	}
}

func TestGoodPacketMarshalBinary(t *testing.T) {
	testCases := []struct {
		name   string
		bytes  []byte
		packet mopp.Packet
	}{
		{
			name:  "PARIS",
			bytes: []byte{0x5b, 0x41, 0xa4, 0x61, 0x91, 0x45, 0x70},
			packet: mopp.Packet{
				ProtoVersion: 0b01,
				SerialNumber: 0b011011,
				Speed:        0b010000,
				States: []morse.State{
					morse.StateDit,
					morse.StateDah,
					morse.StateDah,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDah,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDah,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDit,
					morse.StateEOC,

					morse.StateDit,
					morse.StateDit,
					morse.StateDit,
					morse.StateEOW,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.packet.MarshalBinary()
			require.NoError(t, err)
			require.Equal(t, tc.bytes, data)
		})
	}
}

func TestBadPacketMarshalBinary(t *testing.T) {
	testCases := []struct {
		name        string
		packet      mopp.Packet
		errorString string
	}{
		{
			name:        "unsupported protocol version",
			packet:      mopp.Packet{},
			errorString: "unsupported protocol version: 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.packet.MarshalBinary()
			require.Error(t, err)
			require.EqualError(t, err, tc.errorString)
		})
	}
}
