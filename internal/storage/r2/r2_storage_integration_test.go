package r2_test

import (
	"KnowLedger/internal/storage/r2"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationStorage(t *testing.T) *r2.CASStorage {
	t.Helper()

	vars := []string{"R2_PUBLIC_ENDPOINT", "R2_ACCESS_KEY", "R2_SECRET_KEY", "R2_BUCKET_NAME", "R2_API_ENDPOINT"}
	for _, v := range vars {
		if os.Getenv(v) == "" {
			t.Skipf("skipping integration test: %s not set", v)
		}
	}

	s, err := r2.NewR2CASStorage(
		os.Getenv("R2_BUCKET_NAME"),
		os.Getenv("R2_ACCESS_KEY"),
		os.Getenv("R2_SECRET_KEY"),
		os.Getenv("R2_API_ENDPOINT"),
		os.Getenv("R2_PUBLIC_ENDPOINT"),
	)
	require.NoError(t, err)
	return s
}

func TestIntegration_UploadAndExists(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()
	data := []byte("integration test content")

	result, err := s.Upload(ctx, data)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Key)
	assert.NotEmpty(t, result.Hash)
	assert.NotEmpty(t, result.ContentType)
	assert.NotEmpty(t, result.URL)
	assert.Equal(t, int64(len(data)), result.Size)
	t.Logf("public url: %s", result.URL)

	exists, err := s.Exists(ctx, result.Key)
	require.NoError(t, err)
	assert.True(t, exists)

	t.Cleanup(func() { _ = s.Delete(ctx, result.Key) })
}

func TestIntegration_UploadDeduplication(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()
	data := []byte("dedup test content")

	result1, err := s.Upload(ctx, data)
	require.NoError(t, err)

	result2, err := s.Upload(ctx, data)
	require.NoError(t, err)

	assert.Equal(t, result1.Key, result2.Key)
	assert.Equal(t, result1.Hash, result2.Hash)

	t.Cleanup(func() { _ = s.Delete(ctx, result1.Key) })
}

func TestIntegration_DeleteObject(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()
	data := []byte("to be deleted")

	result, err := s.Upload(ctx, data)
	require.NoError(t, err)

	err = s.Delete(ctx, result.Key)
	require.NoError(t, err)

	exists, err := s.Exists(ctx, result.Key)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestIntegration_GetURL(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()
	data := []byte("url test content")

	result, err := s.Upload(ctx, data)
	require.NoError(t, err)
	assert.Contains(t, result.URL, result.Key)

	// Verifikasi GetURL konsisten dengan PublicURL dari Upload
	url := s.GetURL(ctx, result.Key)
	assert.Equal(t, result.URL, url)
	t.Logf("public url: %s", url)

	t.Cleanup(func() { _ = s.Delete(ctx, result.Key) })
}

func TestIntegration_Upload_UnsupportedType(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()

	// Data binary random yang tidak dikenali mimetype-nya
	data := []byte{0x00, 0x01, 0x02, 0x03}

	_, err := s.Upload(ctx, data)
	assert.EqualError(t, err, "unsupported or unrecognized file type")
}

func TestIntegration_GetMediaDetails(t *testing.T) {
	s := newIntegrationStorage(t)
	ctx := context.Background()

	data := []byte("url test content")

	result, err := s.Upload(ctx, data)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Key)

	media, err := s.GetDetails(ctx, result.Key)
	require.NoError(t, err)

	assert.Equal(t, result, media)

	t.Cleanup(func() { _ = s.Delete(ctx, result.Key) })
}
