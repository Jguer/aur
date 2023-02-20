package rpc

import (
	"context"
	"net/http"
	"strings"

	"github.com/Jguer/aur"
)

const _defaultURL = "https://aur.archlinux.org/rpc?"
const defaultBatchSize = 125

type ClientInterface interface {
	aur.QueryClient
	// Search queries the AUR DB with an optional By filter.
	// Use By.None for default query param (name-desc)
	Search(ctx context.Context, query string, by aur.By) ([]aur.Pkg, error)

	// Info gives detailed information on existing package.
	Info(ctx context.Context, pkgs []string) ([]aur.Pkg, error)
}

type LogFn func(a ...any)

// Client for AUR searching and querying.
type Client struct {
	BaseURL string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	HTTPClient aur.HTTPRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []aur.RequestEditorFn

	// Batch size for batch requests.
	batchSize int

	// Log Function for debugging.
	logFn LogFn

	// cache for storing info results
	cache map[string]aur.Pkg
}

// ClientOption allows setting custom parameters during construction.
type ClientOption func(*Client) error

func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		BaseURL:        _defaultURL,
		HTTPClient:     nil,
		RequestEditors: []aur.RequestEditorFn{},
		batchSize:      defaultBatchSize,
		logFn:          nil,
		cache:          make(map[string]aur.Pkg),
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
	if !strings.HasSuffix(client.BaseURL, "rpc?") {
		// ensure the server URL always has a trailing slash
		if !strings.HasSuffix(client.BaseURL, "/") {
			client.BaseURL += "/"
		}

		client.BaseURL += "rpc?"
	}

	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer aur.HTTPRequestDoer) ClientOption {
	return func(c *Client) error {
		c.HTTPClient = doer

		return nil
	}
}

// WithBatchSize allows overriding the default value for batch size.
func WithBatchSize(batchSize int) ClientOption {
	return func(c *Client) error {
		c.batchSize = batchSize

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

// WithLogFn allows overriding the default log function.
func WithLogFn(fn LogFn) ClientOption {
	return func(c *Client) error {
		c.logFn = fn

		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn aur.RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)

		return nil
	}
}
