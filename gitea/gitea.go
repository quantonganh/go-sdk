// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gitea

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-querystring/query"
)

// Version return the library version
func Version() string {
	return "0.12.3"
}

// Client represents a Gogs API client.
type Client struct {
	url         string
	accessToken string
	sudo        string
	client      *http.Client
}

// NewClient initializes and returns a API client.
func NewClient(url, token string) *Client {
	return &Client{
		url:         strings.TrimSuffix(url, "/"),
		accessToken: token,
		client:      &http.Client{},
	}
}

// NewClientWithHTTP creates an API client with a custom http client
func NewClientWithHTTP(url string, httpClient *http.Client) {
	client := NewClient(url, "")
	client.client = httpClient
}

// SetHTTPClient replaces default http.Client with user given one.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.client = client
}

// SetSudo sets username to impersonate.
func (c *Client) SetSudo(sudo string) {
	c.sudo = sudo
}

func (c *Client) doRequest(method, path string, header http.Header, body interface{}) (*http.Response, error) {
	u, err := url.Parse(c.url + "/api/v1" + path)
	if err != nil {
		return nil, err
	}

	if method == "GET" {
		if body != nil {
			q, err := query.Values(body)
			if err != nil {
				return nil, err
			}
			u.RawQuery = q.Encode()
		}
	}

	req := &http.Request{
		Method:     method,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(bodyBytes)

		u.RawQuery = ""
		req.Body = ioutil.NopCloser(bodyReader)
		req.GetBody = func() (io.ReadCloser, error) {
			return ioutil.NopCloser(bodyReader), nil
		}
		req.ContentLength = int64(bodyReader.Len())
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")

	if len(c.accessToken) != 0 {
		req.Header.Set("Authorization", "token "+c.accessToken)
	}
	if c.sudo != "" {
		req.Header.Set("Sudo", c.sudo)
	}
	for k, v := range header {
		req.Header[k] = v
	}

	return c.client.Do(req)
}

func (c *Client) getResponse(method, path string, header http.Header, body interface{}) ([]byte, error) {
	resp, err := c.doRequest(method, path, header, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 403:
		return nil, errors.New("403 Forbidden")
	case 404:
		return nil, errors.New("404 Not Found")
	case 409:
		return nil, errors.New("409 Conflict")
	case 422:
		return nil, fmt.Errorf("422 Unprocessable Entity: %s", string(data))
	}

	if resp.StatusCode/100 != 2 {
		errMap := make(map[string]interface{})
		if err = json.Unmarshal(data, &errMap); err != nil {
			// when the JSON can't be parsed, data was probably empty or a plain string,
			// so we try to return a helpful error anyway
			return nil, fmt.Errorf("Unknown API Error: %d %s", resp.StatusCode, string(data))
		}
		return nil, errors.New(errMap["message"].(string))
	}

	return data, nil
}

func (c *Client) getParsedResponse(method, path string, header http.Header, body interface{}, obj interface{}) error {
	data, err := c.getResponse(method, path, header, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

func (c *Client) getStatusCode(method, path string, header http.Header, body io.Reader) (int, error) {
	resp, err := c.doRequest(method, path, header, body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}
