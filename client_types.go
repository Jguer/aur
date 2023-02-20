package aur

import (
	"context"
	"net/http"
)

type Query struct {
	Needles  []string
	By       By
	Contains bool // if true, search for packages containing the needle, not exact matches
}

// RequestEditorFn  is the function signature for the RequestEditor callback function.
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// HTTPRequestDoer performs HTTP requests.
// The standard http.Client implements this interface.
type HTTPRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// HTTPRequestDoer performs HTTP requests.
// The standard http.Client implements this interface.
type QueryClient interface {
	Get(ctx context.Context, query *Query) ([]Pkg, error)
}
