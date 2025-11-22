//go:build !wasm

package fetchgo

import "encoding/json"

func (f *Fetchgo) getJSONEncoder() encoder { return &stdlibJSONEncoder{} }

// stdlibJSONEncoder encodes Go values to JSON []byte
type stdlibJSONEncoder struct{}

func (e *stdlibJSONEncoder) Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}
