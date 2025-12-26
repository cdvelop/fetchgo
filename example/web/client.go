//go:build wasm

package main

import (
	"syscall/js"

	"github.com/tinywasm/binary"
	"github.com/tinywasm/fetch"
	"github.com/tinywasm/fetch/example/model"
)

func main() {
	// Get the current window location origin to use as base URL
	origin := js.Global().Get("location").Get("origin").String()

	// Setup UI
	document := js.Global().Get("document")
	body := document.Get("body")

	// Title
	h1 := document.Call("createElement", "h1")
	h1.Set("innerHTML", "Fetch Binary Example")
	body.Call("appendChild", h1)

	// Info
	p := document.Call("createElement", "p")
	p.Set("innerHTML", "Click the button to send a User struct (binary) to the server.")
	body.Call("appendChild", p)

	// Button
	btn := document.Call("createElement", "button")
	btn.Set("innerText", "Send User Data")
	body.Call("appendChild", btn)

	// Result container
	resultDiv := document.Call("createElement", "div")
	resultDiv.Set("style", "margin-top: 20px; padding: 10px; border: 1px solid #ccc;")
	body.Call("appendChild", resultDiv)

	// Click handler
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resultDiv.Set("innerText", "Sending...")

		user := model.User{
			Name:  "Alice",
			Email: "alice@example.com",
			Age:   30,
		}

		// Encode user to binary
		var data []byte
		if err := binary.Encode(&user, &data); err != nil {
			resultDiv.Set("innerText", "Encode error: "+err.Error())
			return nil
		}

		// Send POST request with binary data
		fetch.Post(origin + "/api/user").
			ContentTypeBinary().
			Body(data).
			Timeout(5000).
			Send(func(resp *fetch.Response, err error) {
				if err != nil {
					resultDiv.Set("innerText", "Error: "+err.Error())
					return
				}
				resultDiv.Set("innerText", "Response: "+resp.Text())
			})

		return nil
	})
	btn.Call("addEventListener", "click", cb)

	// Keep the program running
	select {}
}
