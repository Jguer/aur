package rpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Jguer/aur"
)

type response struct {
	Error       string    `json:"error"`
	Type        string    `json:"type"`
	Version     int       `json:"version"`
	ResultCount int       `json:"resultcount"`
	Results     []aur.Pkg `json:"results"`
}

func newAURRPCRequest(ctx context.Context, baseURL string, values url.Values) (*http.Request, error) {
	values.Set("v", "5")

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+values.Encode(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return req, nil
}

func parseRPCResponse(resp *http.Response) ([]aur.Pkg, error) {
	defer resp.Body.Close()

	if err := aur.GetErrorByStatusCode(resp.StatusCode); err != nil {
		return nil, err
	}

	result := new(response)

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, fmt.Errorf("response decoding failed: %w", err)
	}

	if len(result.Error) > 0 {
		return nil, &aur.PayloadError{
			StatusCode: resp.StatusCode,
			ErrorField: result.Error,
		}
	}

	return result.Results, nil
}
