package fetchgo

// upsertHeader is an internal helper that implements the core logic for adding or replacing
// a header. If replace==true, it will replace the first occurrence of the key;
// otherwise, it will append a new (key, value) pair, even if the key already exists.
func (c *Client) upsertHeader(key, value string, replace bool) {
	for i := 0; i < len(c.defaultHeaders); i += 2 {
		if i+1 < len(c.defaultHeaders) && c.defaultHeaders[i] == key {
			if replace {
				c.defaultHeaders[i+1] = value
				return
			}
			// If not replacing, add the duplicate and return.
			c.defaultHeaders = append(c.defaultHeaders, key, value)
			return
		}
	}
	// If the key is not found, append it.
	c.defaultHeaders = append(c.defaultHeaders, key, value)
}

// AddHeader adds a header to the client's default headers.
// It will add the (key, value) pair even if the key already exists,
// potentially creating duplicate header keys.
func (c *Client) AddHeader(key, value string) {
	c.upsertHeader(key, value, false)
}

// SetHeader sets a header, ensuring there is at most one entry for the given key.
// If the header already exists, its value is replaced. If it does not exist, it is appended.
func (c *Client) SetHeader(key, value string) {
	c.upsertHeader(key, value, true)
}

// getHeaders converts the private []string representation to a map[string]string for
// internal use by the request executors. If duplicate keys exist in the slice, the
// last value wins, which is consistent with the behavior of SetHeader.
func (c *Client) getHeaders() map[string]string {
	if len(c.defaultHeaders) == 0 {
		return nil
	}
	headers := make(map[string]string, len(c.defaultHeaders)/2)
	for i := 0; i < len(c.defaultHeaders); i += 2 {
		if i+1 < len(c.defaultHeaders) {
			headers[c.defaultHeaders[i]] = c.defaultHeaders[i+1]
		}
	}
	return headers
}
