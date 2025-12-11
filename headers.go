package fetch

// getHeaders returns the default headers as a map for internal use.
func (c *client) getHeaders() map[string]string {
	if len(c.defaultHeaders) == 0 {
		return nil
	}
	// Return a copy to prevent external modification
	headers := make(map[string]string, len(c.defaultHeaders))
	for k, v := range c.defaultHeaders {
		headers[k] = v
	}
	return headers
}
