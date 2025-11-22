package fetchgo

import (
	"encoding/json"

	. "github.com/cdvelop/tinystring"
)

// JSONEncoder implements the encoder interface for JSON data.
type JSONEncoder struct{}

// Encode marshals the given data into a JSON byte slice.
func (e JSONEncoder) Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}

// RawEncoder implements the encoder interface for raw byte data.
// It acts as a pass-through for []byte and converts string to []byte.
type RawEncoder struct{}

// Encode handles raw data. It expects data to be either []byte or string.
func (e RawEncoder) Encode(data any) ([]byte, error) {
	switch d := data.(type) {
	case []byte:
		return d, nil
	case string:
		return []byte(d), nil
	default:
		return nil, Errf("raw encoder: unsupported type %T", data)
	}
}
