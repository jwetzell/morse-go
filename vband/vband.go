package vband

type VBandMessage interface {
	Encode() (string, error)
	Decode(string) error
}
