//go:build wasm

package fetchgo

import (
	. "github.com/cdvelop/tinystring"
	"syscall/js"
)

// doRequest is the WASM implementation for making an HTTP request using the browser's fetch API.
func (c *Client) doRequest(method, endpoint string, body any) (any, error) {
	// done channel will block until the fetch promise resolves or rejects.
	done := make(chan struct {
		result any
		err    error
	}, 1)

	// 1. Build the full URL.
	fullURL, err := c.buildURL(endpoint)
	if err != nil {
		return nil, err
	}

	// 2. Prepare the request body for JavaScript's fetch.
	var jsBody js.Value
	var contentType string
	if body != nil {
		switch b := body.(type) {
		case js.Value:
			jsBody = b // Pass js.Value directly (e.g., for File objects from an <input>).
		case []byte:
			// Convert Go byte slice to a JS Uint8Array's buffer.
			uint8Array := js.Global().Get("Uint8Array").New(len(b))
			js.CopyBytesToJS(uint8Array, b)
			jsBody = uint8Array.Get("buffer")
		case string:
			jsBody = js.ValueOf(b) // Simple strings can be passed directly.
		default:
			// For other types (structs, maps), use the configured encoder.
			encoderToUse := c.encoder
			if encoderToUse == nil {
				switch c.RequestType {
				case RequestJSON:
					encoderToUse = &JSONEncoder{}
					contentType = "application/json; charset=utf-8"
				default:
					encoderToUse = &RawEncoder{}
				}
			}

			encoded, err := encoderToUse.Encode(body)
			if err != nil {
				return nil, Errf("wasm encoding error: %s", err.Error())
			}

			// For JSON, send a string body. For raw, send an ArrayBuffer.
			if c.RequestType == RequestJSON {
				jsBody = js.ValueOf(string(encoded))
			} else {
				uint8Array := js.Global().Get("Uint8Array").New(len(encoded))
				js.CopyBytesToJS(uint8Array, encoded)
				jsBody = uint8Array.Get("buffer")
			}
		}
	}

	// 3. Prepare headers object for the fetch call.
	headers := c.getHeaders()
	jsHeaders := js.Global().Get("Object").New()
	if contentType != "" {
		if headers == nil {
			headers = make(map[string]string)
		}
		if _, exists := headers["Content-Type"]; !exists {
			headers["Content-Type"] = contentType
		}
	}
	for k, v := range headers {
		jsHeaders.Set(k, v)
	}

	// 4. Prepare the main options object for fetch.
	options := js.Global().Get("Object").New()
	options.Set("method", method)
	options.Set("headers", jsHeaders)
	if !jsBody.IsUndefined() {
		options.Set("body", jsBody)
	}

	// 5. Handle timeout with AbortController.
	if c.TimeoutMS > 0 {
		controller := js.Global().Get("AbortController").New()
		options.Set("signal", controller.Get("signal"))
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			controller.Call("abort")
			return nil
		}), c.TimeoutMS)
	}

	// 6. Define promise handlers to bridge async JS to sync Go.
	var success, failure, responseHandler js.Func

	// success handles the final response body (as an ArrayBuffer).
	success = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer success.Release()
		// Convert the JS ArrayBuffer back to a Go []byte.
		uint8Array := js.Global().Get("Uint8Array").New(args[0])
		goBytes := make([]byte, uint8Array.Get("length").Int())
		js.CopyBytesToGo(goBytes, uint8Array)
		done <- struct {
			result any
			err    error
		}{result: goBytes, err: nil}
		return nil
	})

	// failure handles any error in the promise chain.
	failure = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer failure.Release()
		err := Errf("fetch failed: %s", args[0].Call("toString").String())
		done <- struct {
			result any
			err    error
		}{result: nil, err: err}
		return nil
	})

	// responseHandler handles the initial Response object from fetch.
	responseHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		defer responseHandler.Release()
		response := args[0]
		if !response.Get("ok").Bool() {
			// If status is not 2xx, read body as text for the error message.
			response.Call("text").Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				errMsg := args[0].String()
				status := response.Get("status").Int()
				err := Errf("request failed with status %d: %s", status, errMsg)
				done <- struct {
					result any
					err    error
				}{result: nil, err: err}
				return nil
			}))
		} else {
			// On success, get the body as an ArrayBuffer, which also returns a promise.
			response.Call("arrayBuffer").Call("then", success, failure)
		}
		return nil
	})

	// 7. Execute the fetch call and start the promise chain.
	js.Global().Call("fetch", fullURL, options).Call("then", responseHandler, failure)

	// 8. Block and wait for one of the handlers to send a result on the channel.
	res := <-done
	return res.result, res.err
}
