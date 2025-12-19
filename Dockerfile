# Multi-stage build for Go service
FROM golang:1.21-alpine AS builder

WORKDIR /src

# dependencies
COPY go.mod go.sum ./
RUN go mod download

# build
COPY . ./
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags "-s -w" -o /server ./cmd/server

# final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /server /server
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/server"]
