package shortener

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestBase62EncodeDecode(t *testing.T) {
	examples := []int64{0, 1, 10, 61, 62, 12345, 999999}
	for _, n := range examples {
		s := encodeBase62(n)
		m, err := decodeBase62(s)
		require.NoError(t, err)
		require.Equal(t, n, m)
	}
}

func TestShortenResolve(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	svc := New(rdb, "http://localhost:8080")

	ctx := context.Background()
	shortURL, err := svc.Shorten(ctx, "https://example.com/abc")
	require.NoError(t, err)
	require.NotEmpty(t, shortURL)
	// Shorten returns a full short URL (base + id), it should not contain the original URL
	require.Contains(t, shortURL, "http://localhost:8080/")
	orig, err := svc.Resolve(ctx, shortURL[len("http://localhost:8080/"):])
	require.NoError(t, err)
	require.Equal(t, "https://example.com/abc", orig)
}

func TestDeterministicShorten(t *testing.T) {
	srv, err := miniredis.Run()
	require.NoError(t, err)
	defer srv.Close()

	rdb := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	svc := New(rdb, "http://localhost:8080")

	ctx := context.Background()
	shortURL1, err := svc.Shorten(ctx, "https://example.com/foo")
	require.NoError(t, err)
	shortURL2, err := svc.Shorten(ctx, "https://example.com/foo")
	require.NoError(t, err)

	// Should be deterministic: same input -> same short URL
	require.Equal(t, shortURL1, shortURL2)

	id1 := shortURL1[len("http://localhost:8080/"):]
	orig1, err := svc.Resolve(ctx, id1)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/foo", orig1)
}
