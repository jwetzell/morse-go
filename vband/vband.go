package vband

type VBandMessage interface {
	Encode() ([]byte, error)
	Decode([]byte) error
}
