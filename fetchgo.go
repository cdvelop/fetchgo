package fetch

import (
	"github.com/tinywasm/gobin"
	"github.com/tinywasm/json"
)

// Fetch manages HTTP clients with explicit codec methods.
type Fetch struct {
	tb              *gobin.GoBin
	tj              *json.JSON
	corsMode        string
	corsCredentials bool
}

// New creates a new Fetch instance with sensible defaults.
func New() *Fetch {
	return &Fetch{
		tb:              gobin.New(),
		tj:              json.New(),
		corsMode:        "cors",
		corsCredentials: false,
	}
}

// SetCORS configures CORS behavior for WASM/browser requests.
func (f *Fetch) SetCORS(mode string, credentials bool) *Fetch {
	f.corsMode = mode
	f.corsCredentials = credentials
	return f // Chainable
}

// NewClient creates a configured HTTP client.
func (f *Fetch) NewClient(baseURL string, timeoutMS int) Client {
	return &client{
		baseURL:        baseURL,
		timeoutMS:      timeoutMS,
		fch:            f, // Reference to parent
		defaultHeaders: make(map[string]string),
	}
}
