//go:build !wasm

package main

import (
	"compress/gzip"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinywasm/fetch/example/model"
	"github.com/tinywasm/tinybin"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func main() {
	// Define flags
	publicDir := flag.String("public-dir", "", "Directory containing static files")
	port := flag.String("port", "", "Port to listen on")
	flag.Parse()

	// Priority: flag > env var > default
	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = "6060"
		}
	}

	if *publicDir == "" {
		*publicDir = os.Getenv("PUBLIC_DIR")
		if *publicDir == "" {
			*publicDir = "public"
		}
	}

	// Make it absolute if it's relative
	absPublicDir, err := filepath.Abs(*publicDir)
	if err != nil {
		log.Fatalf("Error resolving public directory path: %v", err)
	}

	// Verify the directory exists
	if _, err := os.Stat(absPublicDir); os.IsNotExist(err) {
		log.Fatalf("Static files directory does not exist: %s", absPublicDir)
	}

	log.Printf("Serving static files from: %s", absPublicDir)
	fs := http.FileServer(http.Dir(absPublicDir))

	// Middleware to disable caching for static files (useful in dev/test)
	noCache := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			h.ServeHTTP(w, r)
		})
	}

	mux := http.NewServeMux()

	// Middleware to enable gzip compression
	gzipHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gzw := &gzipResponseWriter{Writer: gz, ResponseWriter: w}
			next.ServeHTTP(gzw, r)
		})
	}

	mux.Handle("/", noCache(gzipHandler(fs)))

	// API endpoint to receive User data via TinyBin
	mux.HandleFunc("/api/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Read error", http.StatusInternalServerError)
			return
		}

		// Decode TinyBin
		var user model.User
		err = tinybin.New().Decode(body, &user)
		if err != nil {
			http.Error(w, "Decode error: "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Received user: %+v", user)

		// Respond
		w.Write([]byte("Server received: " + user.Name))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running"))
	})

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	log.Printf("Starting server on port %s", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
