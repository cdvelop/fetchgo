package fetch

// client is the private implementation of the Client interface.
type client struct {
	baseURL        string
	defaultHeaders map[string]string
	timeoutMS      int
	fch            *Fetch // Reference to parent for codec/config access
}

func (c *client) SendJSON(method, url string, body any, callback func([]byte, error)) {
	var encodedBody []byte
	var err error

	if body != nil {
		encodedBody, err = c.fch.tj.Encode(body)
		if err != nil {
			callback(nil, err)
			return
		}
	}
	c.doRequest(method, url, "application/json; charset=utf-8", encodedBody, callback)
}

// SendBinary encodes body with TinyBin and sends HTTP request.
func (c *client) SendBinary(method, url string, body any, callback func([]byte, error)) {
	var encodedBody []byte
	var err error

	if b, ok := body.([]byte); ok {
		encodedBody = b
	} else {
		encodedBody, err = c.fch.tb.Encode(body)
		if err != nil {
			callback(nil, err)
			return
		}
	}

	c.doRequest(method, url, "application/octet-stream", encodedBody, callback)
}

// SetHeader adds or updates a default header for all requests from this client.
func (c *client) SetHeader(key, value string) {
	c.defaultHeaders[key] = value
}
