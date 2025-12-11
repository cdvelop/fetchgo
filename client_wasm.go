//go:build wasm

package fetch

import (
	"syscall/js"

	. "github.com/tinywasm/fmt"
)

// doRequest is the WASM implementation for making an HTTP request using the browser's fetch API.
func (c *client) doRequest(method, endpoint string, contentType string, body []byte, callback func([]byte, error)) {
	// 1. Build the full URL.
	fullURL, err := c.buildURL(endpoint)
	if err != nil {
		callback(nil, err)
		return
	}

	// 2. Prepare request body.
	var jsBody js.Value
	if len(body) > 0 {
		// Convert Go byte slice to a JS Uint8Array's buffer.
		uint8Array := js.Global().Get("Uint8Array").New(len(body))
		js.CopyBytesToJS(uint8Array, body)
		jsBody = uint8Array.Get("buffer")
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
	if c.timeoutMS > 0 {
		controller := js.Global().Get("AbortController").New()
		options.Set("signal", controller.Get("signal"))
		js.Global().Call("setTimeout", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			controller.Call("abort")
			return nil
		}), c.timeoutMS)
	}

	// 6. Define promise handlers to bridge async JS to sync Go.
	var success, failure, responseHandler js.Func

	// cleanup releases the JS functions when the request is complete.
	cleanup := func() {
		success.Release()
		failure.Release()
		responseHandler.Release()
	}

	// success handles the final response body (as an ArrayBuffer).
	success = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Convert the JS ArrayBuffer back to a Go []byte.
		uint8Array := js.Global().Get("Uint8Array").New(args[0])
		goBytes := make([]byte, uint8Array.Get("length").Int())
		js.CopyBytesToGo(goBytes, uint8Array)

		callback(goBytes, nil)
		cleanup()
		return nil
	})

	// failure handles any error in the promise chain.
	failure = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var errMsg string
		if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
			jsErr := args[0]
			// Try to get error message
			if jsErr.Type() == js.TypeString {
				errMsg = jsErr.String()
			} else if jsErr.Get("message").Type() == js.TypeString {
				errMsg = jsErr.Get("message").String()
			} else {
				errMsg = jsErr.Call("toString").String()
			}
		}
		if errMsg == "" {
			errMsg = "unknown network error (possibly CORS, network unavailable, or invalid URL)"
		}
		err := Errf("fetch failed: %s (URL: %s)", errMsg, fullURL)

		callback(nil, err)
		cleanup()
		return nil
	})

	// responseHandler handles the initial Response object from fetch.
	responseHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		response := args[0]
		status := response.Get("status").Int()

		if !response.Get("ok").Bool() {
			// If status is not 2xx, read body as text for the error message.
			var textHandler js.Func
			textHandler = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				defer textHandler.Release()
				errMsg := args[0].String()
				err := Errf("request failed with status %d: %s", status, errMsg)

				callback(nil, err)
				cleanup()
				return nil
			})
			response.Call("text").Call("then", textHandler, failure)
		} else {
			// On success, get the body as an ArrayBuffer, which also returns a promise.
			response.Call("arrayBuffer").Call("then", success, failure)
		}
		return nil
	})

	// 7. Execute the fetch call and start the promise chain.
	js.Global().Call("fetch", fullURL, options).Call("then", responseHandler, failure)
}
