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
	id, shortURL, err := svc.Shorten(ctx, "https://example.com/abc")
	require.NoError(t, err)
	require.NotEmpty(t, id)
	require.LessOrEqual(t, len(id), 6)
	require.Contains(t, shortURL, id)

	orig, err := svc.Resolve(ctx, id)
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
	id1, _, err := svc.Shorten(ctx, "https://example.com/foo")
	require.NoError(t, err)
	require.LessOrEqual(t, len(id1), 6)
	id2, _, err := svc.Shorten(ctx, "https://example.com/foo")
	require.NoError(t, err)

	orig1, err := svc.Resolve(ctx, id1)
	require.NoError(t, err)
	orig2, err := svc.Resolve(ctx, id2)
	require.NoError(t, err)
	require.Equal(t, "https://example.com/foo", orig1)
	require.Equal(t, "https://example.com/foo", orig2)
}
