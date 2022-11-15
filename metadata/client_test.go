package metadata

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientCreation(t *testing.T) {
	t.Parallel()
	client, err := New()

	require.NoError(t, err)

	assert.Equal(t, baseURL, client.baseURL)
	assert.Equal(t, cacheValidity, client.cacheValidity)
	assert.Equal(t, http.DefaultClient, client.httpClient)
	assert.NotEmpty(t, client.cacheFilePath)
	assert.Nil(t, client.debugLoggerFn)
	assert.Nil(t, client.unmarshalledCache)
}

func TestClientCreationWithCustomOptions(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp(t.TempDir(), "aur-cache-*")
	httpClient := &http.Client{}
	require.NoError(t, err)

	client, err := New(
		WithBaseURL("http://foo.bar"),
		WithCustomCacheValidity(10),
		WithHTTPClient(httpClient),
		WithCacheFilePath(dir+"/cache.json"),
		WithDebugLogger(func(a ...any) {}),
		WithRequestEditorFn(func(ctx context.Context, r *http.Request) error { return nil }),
	)

	require.NoError(t, err)

	assert.Equal(t, "http://foo.bar", client.baseURL)
	assert.Equal(t, time.Duration(10), client.cacheValidity)
	assert.Equal(t, http.DefaultClient, client.httpClient)
	assert.Equal(t, dir+"/cache.json", client.cacheFilePath)
	assert.NotNil(t, client.debugLoggerFn)
	assert.NotNil(t, client.requestEditors)
	assert.Nil(t, client.unmarshalledCache)
}

func TestClientCreationWithInvalidCachePath(t *testing.T) {
	t.Parallel()
	_, err := New(WithCacheFilePath("/foo/bar/baz")) // unexisting path is ok
	assert.NoError(t, err)

	dir := t.TempDir()

	_, err = New(WithCacheFilePath(dir))
	assert.Error(t, err)
}
