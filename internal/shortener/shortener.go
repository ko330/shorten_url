package shortener

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

var (
	ErrNotFound = errors.New("not found")
	alphabet    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type Shortener struct {
	rdb     *redis.Client
	baseURL string
}

func New(rdb *redis.Client, baseURL string) *Shortener {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Shortener{rdb: rdb, baseURL: baseURL}
}

// Shorten stores the original URL and returns the short id derived from the URL's hash.
// It takes the first 48 bits of SHA-256(URL [+ attempt]) and base62-encodes them. On collisions
// (when different URL maps to same id) the function will retry with a small salt.
func (s *Shortener) Shorten(ctx context.Context, original string) (string, error) {
	u, err := url.ParseRequestURI(original)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return "", errors.New("invalid url: must include http:// or https://")
	}

	// Check reverse mapping to avoid duplicates
	revKey := "url:" + original
	if id, err := s.rdb.Get(ctx, revKey).Result(); err == nil {
		return s.baseURL + "/" + id, nil
	}

	// Generate id from URL hash (take first 48 bits). If collision occurs with different URL,
	// retry with a salt (attempt index).
	const maxAttempts = 8
	for i := 0; i < maxAttempts; i++ {
		data := original
		if i > 0 {
			data = original + "#" + strconv.Itoa(i)
		}
		h := sha256.Sum256([]byte(data))
		// take first 5 bytes (40 bits) and convert to uint64 using BigEndian
		var b [8]byte
		copy(b[3:], h[0:5])
		n := int64(binary.BigEndian.Uint64(b[:]))
		id := encodeBase62(n)
		key := "short:" + id

		if _, err := s.rdb.Get(ctx, key).Result(); err == nil {
			// collision with existing short id -> try next attempt
			continue
		} else if err != redis.Nil {
			return "", err
		}

		// key doesn't exist; store it
		if err := s.rdb.Set(ctx, key, original, 0).Err(); err != nil {
			return "", err
		}
		_ = s.rdb.Set(ctx, revKey, id, 0).Err()
		return s.baseURL + "/" + id, nil
	}

	return "", errors.New("failed to generate unique id after retries")
}

func (s *Shortener) Resolve(ctx context.Context, id string) (string, error) {
	key := "short:" + id
	val, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func encodeBase62(num int64) string {
	if num == 0 {
		return string(alphabet[0])
	}
	var b strings.Builder
	for num > 0 {
		rem := num % int64(len(alphabet))
		b.WriteByte(alphabet[rem])
		num = num / int64(len(alphabet))
	}
	// reverse
	runes := []rune(b.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func decodeBase62(s string) (int64, error) {
	var n int64
	for _, r := range s {
		idx := strings.IndexRune(alphabet, r)
		if idx < 0 {
			return 0, errors.New("invalid character in base62")
		}
		n = n*int64(len(alphabet)) + int64(idx)
	}
	return n, nil
}

// Counter removed: ID generation no longer uses a global counter (uses hashed input instead).
