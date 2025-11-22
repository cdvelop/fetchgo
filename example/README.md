# Fetchgo Example

This example demonstrates how to use `fetchgo` to send binary data (TinyBin) from a Go/WASM client to a Go server.

## Structure

- `client.go`: WASM client that uses `fetchgo` to send a `User` struct.
- `server.go`: HTTP server that receives and decodes the `User` struct using `tinybin`.
- `ui/`: Static HTML/CSS files.

## Running the Example

Assuming you have the `golite` environment set up:

1.  Start the development server:
    ```bash
    # From the example/web directory
    golite
    ```

2.  Open your browser at the provided URL (usually `http://localhost:6060`).

3.  Click the "Send User Data" button to trigger the binary request.

## How it works

1.  The client creates a `User` struct.
2.  `fetchgo.SendBinary` automatically encodes the struct using TinyBin.
3.  The browser sends the binary data via HTTP POST.
4.  The server receives the request and decodes the body using `tinybin.Decode`.
5.  The server responds with a text message.
