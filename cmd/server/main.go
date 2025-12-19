package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shorten_url/internal/handler"
	"shorten_url/internal/shortener"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	port := getEnv("PORT", "8080")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	baseURL := getEnv("BASE_URL", "http://localhost:"+port)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis at %s: %v", redisAddr, err)
	}

	svc := shortener.New(rdb, baseURL)
	h := handler.New(svc)

	r := gin.Default()
	r.GET("/health", h.Health)
	r.POST("/api/shorten", h.Shorten)
	r.GET("/:id", h.Resolve)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// graceful shutdown
	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exiting")
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
