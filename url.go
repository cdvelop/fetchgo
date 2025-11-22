package fetchgo

import (
	. "github.com/cdvelop/tinystring"
)

// buildURL constructs the full request URL.
// If the provided endpoint is an absolute URL, it is used as is.
// Otherwise, it is joined with the client's BaseURL.
func (c *client) buildURL(endpoint string) (string, error) {
	// Check if the endpoint is already an absolute URL.
	if HasPrefix(endpoint, "http://") || HasPrefix(endpoint, "https://") {
		return endpoint, nil
	}

	if c.baseURL == "" {
		return "", Err("BaseURL is not set, cannot build URL for relative endpoint")
	}

	// Use PathJoin from tinystring to join the base URL and the endpoint.
	// It handles slash deduplication.
	return PathJoin(c.baseURL, endpoint).String(), nil
}
