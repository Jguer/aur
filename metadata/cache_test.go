package metadata

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientneedsUpdate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cacheFilePath := dir + "/cache.json"

	client, err := New(WithCacheFilePath(cacheFilePath))
	require.NoError(t, err)

	// cache file does not exist
	needs, err := client.needsUpdate()
	require.NoError(t, err)
	require.True(t, needs)

	// cache file exists and is new
	// touch file
	f, err := os.Create(cacheFilePath)
	require.NoError(t, err)
	f.Close()

	needs, err = client.needsUpdate()
	require.NoError(t, err)
	require.False(t, needs)

	// cache file exists and is old
	err = os.Chtimes(cacheFilePath, time.Now().Add(-2*cacheValidity), time.Now().Add(-2*cacheValidity))
	require.NoError(t, err)
	needs, err = client.needsUpdate()
	require.NoError(t, err)
	require.True(t, needs)
}

type MockHTTP struct {
	bytesToReturn []byte
}

func (m *MockHTTP) Do(req *http.Request) (*http.Response, error) {
	body := io.NopCloser(bytes.NewReader(m.bytesToReturn))

	return &http.Response{
		StatusCode: 200,
		Body:       body,
	}, nil
}

func TestClientAssertEndpointCalled(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cacheFilePath := dir + "/cache.json"

	// read test.json
	testBytes, err := os.ReadFile("test.json")
	require.NoError(t, err)

	client, err := New(
		WithCacheFilePath(cacheFilePath),
		WithBaseURL("https://alternative.aur.org"),
		WithHTTPClient(&MockHTTP{bytesToReturn: testBytes}),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			assert.Equal(t, "https://alternative.aur.org/packages-meta-ext-v1.json.gz", req.URL.String())
			req.Header.Add("User-Agent", "test")
			return nil
		},
		))

	require.NoError(t, err)
	ctx := context.Background()

	// cache file does not exist
	_, err = client.makeCache(ctx)
	require.NoError(t, err)
}

func TestClientMakeCache(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cacheFilePath := dir + "/cache.json"

	// read test.json
	testBytes, err := os.ReadFile("test.json")
	require.NoError(t, err)

	client, err := New(
		WithCacheFilePath(cacheFilePath),
		WithHTTPClient(&MockHTTP{bytesToReturn: testBytes}),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", "test")
			return nil
		},
		))
	require.NoError(t, err)

	ctx := context.Background()

	// cache file does not exist
	byNew, err := client.makeCache(ctx)
	require.NoError(t, err)

	assert.Equal(t, testBytes, byNew)

	readCache, err := readCache(cacheFilePath)
	require.NoError(t, err)

	assert.Equal(t, testBytes, readCache)
}

func TestClientCacheAccess(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cacheFilePath := dir + "/cache.json"
	logged := []any{}

	// read test.json
	testBytes, err := os.ReadFile("test.json")
	require.NoError(t, err)

	client, err := New(
		WithCacheFilePath(cacheFilePath),
		WithHTTPClient(&MockHTTP{bytesToReturn: testBytes}),
		WithDebugLogger(func(s ...any) {
			logged = append(logged, s...)
			t.Log(s)
		}),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Add("User-Agent", "test")
			return nil
		},
		))
	require.NoError(t, err)

	ctx := context.Background()

	// cache file does not exist
	cache, err := client.cache(ctx)
	require.NoError(t, err)

	assert.NotNil(t, cache)
	assert.Equal(t, cache, client.unmarshalledCache)
	assert.Equal(t, 1, len(logged))

	// cache is in memory
	cache, err = client.cache(ctx)
	require.NoError(t, err)
	assert.Equal(t, cache, client.unmarshalledCache)
	assert.Equal(t, 1, len(logged))

	// cache file exists
	client.unmarshalledCache = nil
	cache, err = client.cache(ctx)
	require.NoError(t, err)
	assert.Equal(t, cache, client.unmarshalledCache)
	assert.Equal(t, 1, len(logged))
}
