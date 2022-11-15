package metadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/ohler55/ojg/oj"
)

const endpoint = "packages-meta-ext-v1.json.gz"

type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// needsUpdate checks if cachepath is older than 24 hours.
func (a *Client) needsUpdate() (bool, error) {
	// check if cache is older than 24 hours
	info, err := os.Stat(a.cacheFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, fmt.Errorf("unable to read cache: %w", err)
	}

	return info.ModTime().Before(time.Now().Add(-a.cacheValidity)), nil
}

func (a *Client) cache(ctx context.Context) ([]any, error) {
	if a.unmarshalledCache != nil {
		return a.unmarshalledCache, nil
	}

	update, err := a.needsUpdate()
	if err != nil {
		return nil, err
	}

	if update {
		if a.debugLoggerFn != nil {
			a.debugLoggerFn("AUR Cache is out of date, updating")
		}
		cache, makeErr := a.makeCache(ctx)
		if makeErr != nil {
			return nil, makeErr
		}

		inputStruct, unmarshallErr := oj.Parse(cache)
		if unmarshallErr != nil {
			return nil, fmt.Errorf("aur metadata unable to parse cache: %w", unmarshallErr)
		}

		a.unmarshalledCache = inputStruct.([]any)
	} else {
		aurCache, err := readCache(a.cacheFilePath)
		if err != nil {
			return nil, err
		}

		inputStruct, err := oj.Parse(aurCache)
		if err != nil {
			return nil, fmt.Errorf("aur metadata unable to parse cache: %w", err)
		}

		a.unmarshalledCache = inputStruct.([]any)
	}

	return a.unmarshalledCache, nil
}

func readCache(cachePath string) ([]byte, error) {
	fp, err := os.Open(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	defer fp.Close()

	s, err := io.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Download the metadata for aur packages.
// create cache file
// write to cache file.
func (a *Client) makeCache(ctx context.Context) ([]byte, error) {
	body, err := a.downloadAURMetadata(ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	s, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(a.cacheFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err = f.Write(s); err != nil {
		return nil, err
	}

	return s, err
}

func (a *Client) applyEditors(ctx context.Context, req *http.Request) error {
	for _, r := range a.requestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func (a *Client) downloadAURMetadata(ctx context.Context) (io.ReadCloser, error) {
	reqURL, err := url.JoinPath(a.baseURL, endpoint)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if errE := a.applyEditors(ctx, req); errE != nil {
		return nil, errE
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download metadata: %s", resp.Status)
	}

	return resp.Body, nil
}
