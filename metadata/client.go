package metadata

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	cacheValidity = time.Hour
	baseURL       = "https://aur.archlinux.org"
)

type AURCacheClient struct {
	baseURL       string
	cacheValidity time.Duration

	httpClient    HTTPRequestDoer
	cachePath     string
	DebugLoggerFn func(a ...interface{})

	unmarshalledCache []interface{}
}

// ClientOption allows setting custom parameters during construction.
type ClientOption func(*AURCacheClient) error

func New(httpClient HTTPRequestDoer, cachePath string, opts ...ClientOption) (*AURCacheClient, error) {
	client := &AURCacheClient{
		baseURL:       baseURL,
		cacheValidity: cacheValidity,
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

	if client.cachePath == "" {
		dir, err := os.MkdirTemp("", "aur-cache-*")
		if err != nil {
			return nil, fmt.Errorf("aur cache unable to create temp dir: %w", err)
		}

		client.cachePath = dir
	}

	return client, nil
}
