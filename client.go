package aur

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ErrServiceUnavailable represents a error when AUR is unavailable.
var ErrServiceUnavailable = errors.New("AUR is unavailable at this moment")

type PayloadError struct {
	StatusCode int
	ErrorField string
}

func (r *PayloadError) Error() string {
	return fmt.Sprintf("status %d: %s", r.StatusCode, r.ErrorField)
}

const _defaultURL = "https://aur.archlinux.org/rpc/?"

// ClientInterface specification for the AUR client.
type ClientInterface interface {
	// Search queries the AUR DB with an optional By filter.
	// Use By.None for default query param (name-desc)
	Search(ctx context.Context, query string, by By, reqEditors ...RequestEditorFn) ([]Pkg, error)

	// Info gives detailed information on existing package.
	Info(ctx context.Context, pkgs []string, reqEditors ...RequestEditorFn) ([]Pkg, error)
}

// Client for AUR searching and querying.
type Client struct {
	BaseURL string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	HTTPClient HTTPRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction.
type ClientOption func(*Client) error

// HTTPRequestDoer performs HTTP requests.
// The standard http.Client implements this interface.
type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// RequestEditorFn  is the function signature for the RequestEditor callback function.
type RequestEditorFn func(ctx context.Context, req *http.Request) error

func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		BaseURL:        _defaultURL,
		HTTPClient:     nil,
		RequestEditors: []RequestEditorFn{},
	}

	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}

	// create httpClient, if not already present
	if client.HTTPClient == nil {
		client.HTTPClient = http.DefaultClient
	}

	// ensure base URL has /rpc?
	if !strings.HasSuffix(client.BaseURL, "rpc/?") {
		// ensure the server URL always has a trailing slash
		if !strings.HasSuffix(client.BaseURL, "/") {
			client.BaseURL += "/"
		}

		client.BaseURL += "rpc/?"
	}

	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HTTPRequestDoer) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = doer

		return nil
	}
}

// WithBaseURL allows overriding the default base URL of the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		c.BaseURL = baseURL

		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)

		return nil
	}
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}

	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func newAURRPCRequest(ctx context.Context, baseURL string, values url.Values) (*http.Request, error) {
	values.Set("v", "5")

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+values.Encode(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func getErrorByStatusCode(code int) error {
	switch code {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable:
		return ErrServiceUnavailable
	}

	return nil
}

func parseRPCResponse(resp *http.Response) ([]Pkg, error) {
	defer resp.Body.Close()

	if err := getErrorByStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	result := new(response)

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("response decoding failed: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, &PayloadError{
			StatusCode: resp.StatusCode,
			ErrorField: result.Error,
		}
	}

	return result.Results, nil
}

// Search queries the AUR DB with an optional By field.
// Use By.None for default query param (name-desc).
func (c *Client) Search(ctx context.Context, query string, by By, reqEditors ...RequestEditorFn) ([]Pkg, error) {
	values := url.Values{"type": []string{"search"}, "arg": []string{query}}

	if by != None {
		values.Set("by", by.String())
	}

	return c.get(ctx, values, reqEditors)
}

// Info shows info for one or multiple packages.
func (c *Client) Info(ctx context.Context, pkgs []string, reqEditors ...RequestEditorFn) ([]Pkg, error) {
	v := url.Values{"type": []string{"info"}, "arg[]": pkgs}

	return c.get(ctx, v, reqEditors)
}

func (c *Client) get(ctx context.Context, values url.Values, reqEditors []RequestEditorFn) ([]Pkg, error) {
	req, err := newAURRPCRequest(ctx, c.BaseURL, values)
	if err != nil {
		return nil, err
	}

	if errApply := c.applyEditors(ctx, req, reqEditors); errApply != nil {
		return nil, errApply
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return parseRPCResponse(resp)
}
