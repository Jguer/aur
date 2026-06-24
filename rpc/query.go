package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/Jguer/aur"
)

// Search queries the AUR DB with an optional By field.
// Use By.None for default query param (name-desc).
func (c *Client) Search(ctx context.Context, query string, by aur.By) ([]aur.Pkg, error) {
	values := url.Values{"type": []string{"search"}, "arg": []string{query}}

	if by != aur.None {
		values.Set("by", by.String())
	}

	return c.get(ctx, values)
}

// batchSearch queries by each term in the arguments and returns the results aggregated.
func (c *Client) batchSearch(ctx context.Context, queries []string, by aur.By) ([]aur.Pkg, error) {
	pkgs := make([]aur.Pkg, 0, len(queries))
	var err error

	for _, query := range queries {
		tmpPkgs, errS := c.Search(ctx, query, by)
		if errS != nil {
			err = errors.Join(err, errS)
			continue
		}

		pkgs = append(pkgs, tmpPkgs...)
	}

	return pkgs, err
}

// Info shows Info for one or multiple packages.
func (c *Client) Info(ctx context.Context, pkgs []string) ([]aur.Pkg, error) {
	v := url.Values{"type": []string{"info"}, "arg[]": pkgs}

	return c.get(ctx, v)
}

func (c *Client) batchInfo(ctx context.Context, names []string) ([]aur.Pkg, error) {
	info := make([]aur.Pkg, 0, len(names))
	var err error

	missing := make([]string, 0, len(names))
	for _, name := range names {
		if pkg, ok := c.cache[name]; ok {
			info = append(info, pkg)
		} else {
			missing = append(missing, name)
		}
	}

	for n := 0; n < len(missing); n += c.batchSize {
		maxIdx := min(len(missing), n+c.batchSize)

		if c.logFn != nil {
			c.logFn("packages to query", missing[n:maxIdx])
		}

		tempInfo, requestErr := c.Info(ctx, missing[n:maxIdx])
		if requestErr != nil {
			err = errors.Join(err, requestErr)
			continue
		}

		for i := range tempInfo {
			c.cache[tempInfo[i].Name] = tempInfo[i]
		}

		info = append(info, tempInfo...)
	}

	return info, err
}

func (c *Client) get(ctx context.Context, values url.Values) ([]aur.Pkg, error) {
	req, err := newAURRPCRequest(ctx, c.BaseURL, values)
	if err != nil {
		return nil, err
	}

	for _, r := range c.RequestEditors {
		if errR := r(ctx, req); errR != nil {
			return nil, errR
		}
	}

	if c.logFn != nil {
		c.logFn("rpc request", req.URL.String())
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return parseRPCResponse(resp)
}

func (c *Client) Get(ctx context.Context, query *aur.Query) ([]aur.Pkg, error) {
	if len(query.Needles) == 0 {
		return []aur.Pkg{}, nil
	}

	if query.Contains {
		pkgs, err := c.batchSearch(ctx, query.Needles, query.By)
		if err != nil {
			return nil, err
		}

		if c.batchSize != 0 && len(pkgs) < c.batchSize*4 {
			names := make([]string, 0, len(pkgs))
			for i := range pkgs {
				names = append(names, pkgs[i].Name)
			}

			return c.batchInfo(ctx, names)
		}

		return pkgs, nil
	}

	return c.batchInfo(ctx, query.Needles)
}
