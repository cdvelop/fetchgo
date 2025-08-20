package fetchgo

import (
	"encoding/json"
	. "github.com/cdvelop/tinystring"
)

// JSONEncoder implements the Encoder interface for JSON data.
type JSONEncoder struct{}

// Encode marshals the given data into a JSON byte slice.
func (e JSONEncoder) Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}

// Decode unmarshals the given JSON byte slice into the provided variable.
func (e JSONEncoder) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// RawEncoder implements the Encoder interface for raw byte data.
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

// Decode handles raw data. It expects the destination `v` to be a pointer
// to a byte slice (*[]byte).
func (e RawEncoder) Decode(data []byte, v any) error {
	switch d := v.(type) {
	case *[]byte:
		*d = data
		return nil
	default:
		return Errf("raw decoder: unsupported destination type %T, must be *[]byte", v)
	}
}
