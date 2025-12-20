package metadata

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationRealAURDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := t.TempDir()
	cacheFilePath := dir + "/integration_cache.json"

	// Create client with default HTTP client (real network)
	client, err := New(WithCacheFilePath(cacheFilePath))
	require.NoError(t, err)

	ctx := context.Background()

	// This should download the .gz file, transparently decompress it, and save it as JSON
	data, err := client.makeCache(ctx)
	require.NoError(t, err, "makeCache failed")
	assert.NotEmpty(t, data, "Returned data should not be empty")

	// Verify the returned data is valid JSON
	var js interface{}
	err = json.Unmarshal(data, &js)
	assert.NoError(t, err, "Downloaded data should be valid JSON")

	// Verify the file on disk is also valid JSON
	fileData, err := os.ReadFile(cacheFilePath)
	require.NoError(t, err)
	err = json.Unmarshal(fileData, &js)
	assert.NoError(t, err, "Cache file content should be valid JSON")
	assert.NotZero(t, len(fileData), "Cache file content should not be empty")
}
