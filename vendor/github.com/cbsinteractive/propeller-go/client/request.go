package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// get request
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodGet, path, result, nil)
}

// post request
func (c *Client) post(ctx context.Context, path string, result, reqBody interface{}) error {
	return c.request(ctx, http.MethodPost, path, result, reqBody)
}

// put request
func (c *Client) put(ctx context.Context, path string, result, reqBody interface{}) error {
	return c.request(ctx, http.MethodPut, path, result, reqBody)
}

// delete request
func (c *Client) delete(ctx context.Context, path string, result interface{}) error {
	return c.request(ctx, http.MethodDelete, path, result, nil)
}

// request retrieves the resource at url using the provided http method. the response
// body is read to completion, closed, and returned from this function as a byte slice.
func (c *Client) request(parentCtx context.Context, method, path string, result, reqBody interface{}) error {
	var req *http.Request
	var err error

	c.ensure()

	u := c.HostURL.String() + path
	if reqBody != nil {
		body := new(bytes.Buffer)
		err := json.NewEncoder(body).Encode(reqBody)
		if err != nil {
			return err
		}

		req, err = http.NewRequest(method, u, body)
		req.Header.Add("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, u, nil)
	}

	// handle NewRequest error
	if err != nil {
		return err
	}

	c.Auth.Apply(req)

	ctx, cancel := context.WithTimeout(parentCtx, c.Timeout)
	defer cancel()

	resp, err := c.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 > 3 {
		return StatusError{resp.StatusCode, resp.Status, string(body)}
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}

	return nil
}

func (c *Client) ensure() {
	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second
	}
}
