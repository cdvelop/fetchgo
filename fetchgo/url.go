package fetchgo

import (
	"net/url"
	"strings"

	. "github.com/cdvelop/tinystring"
)

// buildURL constructs the full request URL.
// If the provided endpoint is an absolute URL, it is used as is.
// Otherwise, it is joined with the client's BaseURL.
func (c *Client) buildURL(endpoint string) (string, error) {
	// Check if the endpoint is already an absolute URL.
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		// Also validate that it's a well-formed URL.
		if _, err := url.ParseRequestURI(endpoint); err == nil {
			return endpoint, nil
		}
	}

	if c.BaseURL == "" {
		return "", Err("BaseURL is not set, cannot build URL for relative endpoint")
	}

	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", Errf("failed to parse BaseURL: %s", err.Error())
	}

	rel, err := url.Parse(endpoint)
	if err != nil {
		return "", Errf("failed to parse endpoint: %s", err.Error())
	}

	// ResolveReference is the correct way to merge a base URL with a (possibly relative) endpoint.
	return base.ResolveReference(rel).String(), nil
}
