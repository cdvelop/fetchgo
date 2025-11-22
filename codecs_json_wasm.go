//go:build wasm

package fetchgo

import (
	"syscall/js"

	. "github.com/cdvelop/tinystring"
)

func (f *Fetchgo) getJSONEncoder() encoder { return &wasmJSONEncoder{} }

// wasmJSONEncoder uses browser's JSON.stringify via JavaScript API
type wasmJSONEncoder struct{}

func (e *wasmJSONEncoder) Encode(data any) ([]byte, error) {
	// Convert Go value to JavaScript value
	jsValue := convertGoToJS(data)

	// Use browser's JSON.stringify
	jsonString := js.Global().Get("JSON").Call("stringify", jsValue).String()

	return []byte(jsonString), nil
}

// convertGoToJS converts Go values to JavaScript values recursively
func convertGoToJS(data any) js.Value {
	if data == nil {
		return js.Null()
	}

	switch v := data.(type) {
	case string:
		return js.ValueOf(v)
	case bool:
		return js.ValueOf(v)
	case int:
		return js.ValueOf(v)
	case int8:
		return js.ValueOf(int(v))
	case int16:
		return js.ValueOf(int(v))
	case int32:
		return js.ValueOf(int(v))
	case int64:
		return js.ValueOf(int(v))
	case uint:
		return js.ValueOf(int(v))
	case uint8:
		return js.ValueOf(int(v))
	case uint16:
		return js.ValueOf(int(v))
	case uint32:
		return js.ValueOf(int(v))
	case uint64:
		return js.ValueOf(int(v))
	case float32:
		return js.ValueOf(float64(v))
	case float64:
		return js.ValueOf(v)
	case []byte:
		return js.ValueOf(string(v))
	case []any:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, convertGoToJS(item))
		}
		return arr
	case map[string]any:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, convertGoToJS(val))
		}
		return obj
	case map[string]string:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case map[string]int:
		obj := js.Global().Get("Object").New()
		for key, val := range v {
			obj.Set(key, js.ValueOf(val))
		}
		return obj
	case []string:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	case []int:
		arr := js.Global().Get("Array").New(len(v))
		for i, item := range v {
			arr.SetIndex(i, js.ValueOf(item))
		}
		return arr
	default:
		// For other types, try to convert to string using tinystring
		return js.ValueOf(Convert(v).String())
	}
}
