package fetchgo

// encoder defines the interface for encoding and decoding data.
// This allows for pluggable serialization strategies (e.g., JSON, form, raw).
type encoder interface {
	Encode(data any) ([]byte, error)
	Decode(data []byte, v any) error
}

// requestType is a private type that defines how the request body should be
// encoded and which Content-Type header to set. The constants are exported
// so callers can select request kinds without depending on a public type.
type requestType string

const (
	// RequestJSON indicates that the request body should be encoded as JSON.
	RequestJSON requestType = "json"
	// RequestForm indicates that the request body should be form-urlencoded.
	RequestForm requestType = "form"
	// RequestMultipart indicates that the request body should be multipart/form-data.
	RequestMultipart requestType = "multipart"
	// RequestRaw indicates that the request body should be passed as-is.
	RequestRaw requestType = "raw"
)
