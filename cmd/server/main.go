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
	"kabanchik.pro/internal/repo"
	"kabanchik.pro/internal/service"
)

const (
	defaultPort      = "8080"
	defaultMongoURI  = "mongodb://localhost:27017"
	defaultDBName    = "kabanchik"
	defaultJWTSecret = "dev-secret-change"
	defaultJWTTTL    = "24h"
)

func main() {
	port := getEnv("PORT", defaultPort)
	mongoURI := getEnv("MONGO_URI", defaultMongoURI)
	dbName := getEnv("DB_NAME", defaultDBName)
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)
	jwtTTL := getEnv("JWT_TTL", defaultJWTTTL)

	client, err := db.Connect(context.Background(), mongoURI)
	if err != nil {
		log.Fatalf("mongo connect failed: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = client.Disconnect(ctx)
	}()

	database := client.Database(dbName)
	if err := db.EnsureIndexes(context.Background(), database); err != nil {
		log.Fatalf("mongo indexes failed: %v", err)
	}

	store := repo.NewMongoStore(database)
	svc := service.New(store)
	ttl, err := time.ParseDuration(jwtTTL)
	if err != nil {
		log.Fatalf("invalid JWT_TTL: %v", err)
	}
	api := handlers.NewAPI(svc, []byte(jwtSecret), ttl)

	mux := http.NewServeMux()
	api.Register(mux)
	mux.Handle("/", http.FileServer(http.Dir(".")))

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
