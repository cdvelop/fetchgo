package fetch

import (
	"github.com/tinywasm/tinybin"
	"github.com/tinywasm/json"
)

// Fetchgo manages HTTP clients with explicit codec methods.
type Fetchgo struct {
	tb              *tinybin.TinyBin
	tj              *json.TinyJSON
	corsMode        string
	corsCredentials bool
}

// New creates a new Fetchgo instance with sensible defaults.
func New() *Fetchgo {
	return &Fetchgo{
		tb:              tinybin.New(),
		tj:              json.New(),
		corsMode:        "cors",
		corsCredentials: false,
	}
}

// SetCORS configures CORS behavior for WASM/browser requests.
func (f *Fetchgo) SetCORS(mode string, credentials bool) *Fetchgo {
	f.corsMode = mode
	f.corsCredentials = credentials
	return f // Chainable
}

// NewClient creates a configured HTTP client.
func (f *Fetchgo) NewClient(baseURL string, timeoutMS int) Client {
	return &client{
		baseURL:        baseURL,
		timeoutMS:      timeoutMS,
		fetchgo:        f, // Reference to parent
		defaultHeaders: make(map[string]string),
	}
}
