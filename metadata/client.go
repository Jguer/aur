package metadata

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Jguer/aur"
)

const (
	cacheValidity = time.Hour
	baseURL       = "https://aur.archlinux.org"
)

type Client struct {
	baseURL        string
	cacheValidity  time.Duration
	requestEditors []aur.RequestEditorFn
	httpClient     HTTPRequestDoer
	cacheFilePath  string
	debugLoggerFn  LogFn

	unmarshalledCache []any
}

// ClientOption allows setting custom parameters during construction.
type ClientOption func(*Client) error
type LogFn func(a ...any)

func New(opts ...ClientOption) (*Client, error) {
	client := &Client{
		baseURL:           baseURL,
		cacheValidity:     cacheValidity,
		requestEditors:    []aur.RequestEditorFn{},
		httpClient:        nil,
		cacheFilePath:     "",
		debugLoggerFn:     nil,
		unmarshalledCache: nil,
	}

	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(client); err != nil {
			return nil, err
		}
	}

	// create httpClient, if not already present
	if client.httpClient == nil {
		client.httpClient = http.DefaultClient
	}

	if client.cacheFilePath == "" {
		dir, err := os.MkdirTemp("", "aur-cache-*")
		if err != nil {
			return nil, fmt.Errorf("aur cache unable to create temp dir: %w", err)
		}

		client.cacheFilePath = path.Join(dir, "aur-cache.json")
	}

	return client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HTTPRequestDoer) ClientOption {
	return func(c *Client) error {
		c.httpClient = doer

		return nil
	}
}

func WithCustomCacheValidity(duration time.Duration) ClientOption {
	return func(c *Client) error {
		c.cacheValidity = duration

		return nil
	}
}

func WithCacheFilePath(cacheFilePath string) ClientOption {
	return func(c *Client) error {
		info, err := os.Stat(cacheFilePath)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("unable to read cache: %w", err)
		}

		if info != nil && info.IsDir() {
			return fmt.Errorf("cache file path can't be a directory")
		}

		c.cacheFilePath = cacheFilePath

		return nil
	}
}

func WithDebugLogger(logFn LogFn) ClientOption {
	return func(c *Client) error {
		c.debugLoggerFn = logFn

		return nil
	}
}

// WithBaseURL allows overriding the default base URL of the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		c.baseURL = baseURL

		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn aur.RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.requestEditors = append(c.requestEditors, fn)

		return nil
	}
}
