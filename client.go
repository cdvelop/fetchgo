package fetchgo

// client is the private implementation of the Client interface.
type client struct {
	baseURL        string
	defaultHeaders map[string]string
	timeoutMS      int
	fetchgo        *Fetchgo // Reference to parent for codec/config access
}

// SendJSON encodes body as JSON and sends HTTP request.
func (c *client) SendJSON(method, url string, body any, callback func([]byte, error)) {
	encoder := c.fetchgo.getJSONEncoder()
	c.doRequest(method, url, "application/json; charset=utf-8", encoder, body, callback)
}

// SendBinary encodes body with TinyBin and sends HTTP request.
func (c *client) SendBinary(method, url string, body any, callback func([]byte, error)) {
	encoder := c.fetchgo.getTinyBinEncoder()
	c.doRequest(method, url, "application/octet-stream", encoder, body, callback)
}

// SetHeader adds or updates a default header for all requests from this client.
func (c *client) SetHeader(key, value string) {
	c.defaultHeaders[key] = value
}
