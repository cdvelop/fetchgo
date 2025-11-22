package fetchgo

import "github.com/cdvelop/tinybin"

// tinyBinEncoder encodes data using TinyBin
type tinyBinEncoder struct {
	tb *tinybin.TinyBin
}

func (e *tinyBinEncoder) Encode(data any) ([]byte, error) {
	// If data is already []byte, treat it as raw binary data
	if b, ok := data.([]byte); ok {
		return b, nil
	}
	return e.tb.Encode(data)
}
