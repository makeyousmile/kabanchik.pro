package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kabanchik.pro/internal/db"
	"kabanchik.pro/internal/handlers"
)

const (
	defaultPort    = "8080"
	defaultMongoURI = "mongodb://localhost:27017"
	defaultDBName  = "kabanchik"
)

func main() {
	port := getEnv("PORT", defaultPort)
	mongoURI := getEnv("MONGO_URI", defaultMongoURI)
	_ = getEnv("DB_NAME", defaultDBName)

	client, err := db.Connect(context.Background(), mongoURI)
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	mux := http.NewServeMux()
	handlers.Register(mux)
	mux.Handle("/", http.FileServer(http.Dir("web")))

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
