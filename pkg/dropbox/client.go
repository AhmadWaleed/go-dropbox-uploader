package dropbox

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// Client implements dropbox client
type Client struct {
	HTTPClient  *http.Client
	AccessToken string
	Dropbox     *Dropbox
}

// New client with access token
func New(accessToken string, opt UploadOptions) *Client {
	c := &Client{
		HTTPClient:  http.DefaultClient,
		AccessToken: accessToken,
	}

	c.Dropbox = &Dropbox{Client: c, Options: &opt}

	return c
}

func (c *Client) call(path string, opt interface{}) (io.ReadCloser, error) {
	url := "https://api.dropboxapi.com/2" + path

	body, err := json.Marshal(opt)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Content-Type", "application/octet-stream")

	r, err := c.do(req)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) upload(path string, in interface{}, r io.Reader) (io.ReadCloser, error) {
	url := "https://content.dropboxapi.com/2" + path

	body, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("Dropbox-API-Arg", string(body))
	req.Header.Set("Content-Type", "application/octet-stream")

	res, err := c.do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Error bad request
type Error struct {
	Status     string
	StatusCode int
	Summary    string `json:"error_summary"`
}

// Error string.
func (e *Error) Error() string {
	return e.Summary
}

func (c *Client) do(req *http.Request) (io.ReadCloser, error) {
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < http.StatusBadRequest {
		return res.Body, nil
	}

	defer res.Body.Close()

	e := &Error{
		StatusCode: res.StatusCode,
		Status:     http.StatusText(res.StatusCode),
	}

	kind := res.Header.Get("Content-Type")
	if strings.Contains(kind, "text/plain") {
		if b, err := ioutil.ReadAll(res.Body); err == nil {
			e.Summary = string(b)
			return nil, e
		}

		return nil, err
	}

	if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
		return nil, err
	}

	return nil, e
}
