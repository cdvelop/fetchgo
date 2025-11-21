# fetchgo Test Server

Para ejecutar los tests WASM necesitas tener un servidor HTTP corriendo.

## Opción 1: Servidor de prueba standalone

Crea un archivo `testserver/main.go`:

```go
//go:build !wasm

package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func main() {
	mux := http.NewServeMux()

	// Handler for simple GET requests
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method != http.MethodGet {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("get success"))
	})

	// Handler for JSON POST requests, echoes the body back
	mux.HandleFunc("/post_json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		if r.Method == http.MethodOptions {
			return
		}
		
		if r.Header.Get("Content-Type") != "application/json; charset=utf-8" {
			http.Error(w, "bad content type", http.StatusBadRequest)
			return
		}
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Handler for file uploads
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	})

	// Handler that simulates a slow response
	mux.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("slow response"))
	})

	// Handler that always returns an error status
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	log.Println("Test server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Ejecuta:
```bash
cd testserver
go run main.go
```

## Opción 2: Usar test.sh modificado

El script `test.sh` puede iniciar y detener el servidor automáticamente.

## Nota sobre WASM tests

Los tests WASM requieren:
1. Servidor HTTP corriendo (para hacer fetch requests reales)
2. Navegador con soporte WASM (o wasmbrowsertest)
3. CORS habilitado en el servidor

Para ejecutar tests WASM necesitas `wasmbrowsertest`:
```bash
go install github.com/agnivade/wasmbrowsertest@latest
export GOOS=js GOARCH=wasm
go test -v
```