# Shorten URL Service (Gin + Redis)

A simple URL shortening backend using Gin and Redis (no database required). Redis should be available at `localhost:6379` by default.

## Endpoints

- POST /api/shorten
  - Body: `{ "url": "https://example.com" }`
  - Response: `{ "id": "b", "short_url": "http://localhost:8080/b", "url": "https://example.com" }`
- GET /:id
  - Redirects (302) to original URL
- GET /health
  - Returns `{ "status": "ok" }`

## Run

- Ensure Redis is running on `localhost:6379`.
- Build & run:

```bash
# from project root
go run ./cmd/server
```

Or set `REDIS_ADDR`, `PORT`, and `BASE_URL` environment variables to customize.
