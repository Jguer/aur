package metadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/ohler55/ojg/oj"
)

const endpoint = "packages-meta-ext-v1.json.gz"

type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// needsUpdate checks if cachepath is older than 24 hours.
func (a *AURCacheClient) needsUpdate() (bool, error) {
	// check if cache is older than 24 hours
	info, err := os.Stat(a.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}

		return false, fmt.Errorf("unable to read cache: %w", err)
	}

	return info.ModTime().Before(time.Now().Add(-cacheValidity)), nil
}

func (a *AURCacheClient) cache(ctx context.Context) ([]interface{}, error) {
	if a.unmarshalledCache != nil {
		return a.unmarshalledCache, nil
	}

	update, err := a.needsUpdate()
	if err != nil {
		return nil, err
	}

	if update {
		if a.DebugLoggerFn != nil {
			a.DebugLoggerFn("AUR Cache is out of date, updating")
		}
		cache, makeErr := a.makeCache(ctx)
		if makeErr != nil {
			return nil, makeErr
		}

		inputStruct, unmarshallErr := oj.Parse(cache)
		if unmarshallErr != nil {
			return nil, fmt.Errorf("aur metadata unable to parse cache: %w", unmarshallErr)
		}

		a.unmarshalledCache = inputStruct.([]interface{})
	} else {
		aurCache, err := readCache(a.cachePath)
		if err != nil {
			return nil, err
		}

		inputStruct, err := oj.Parse(aurCache)
		if err != nil {
			return nil, fmt.Errorf("aur metadata unable to parse cache: %w", err)
		}

		a.unmarshalledCache = inputStruct.([]interface{})
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
func (a *AURCacheClient) makeCache(ctx context.Context) ([]byte, error) {
	body, err := a.downloadAURMetadata(ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	s, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(a.cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err = f.Write(s); err != nil {
		return nil, err
	}

	return s, err
}

func (a *AURCacheClient) downloadAURMetadata(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", path.Join(a.baseURL, endpoint), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
