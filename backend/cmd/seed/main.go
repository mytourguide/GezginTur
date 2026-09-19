package main

// Admin kullanicisini olusturan basit seed araci.
// Kullanim: go run ./cmd/seed  (ADMIN_EMAIL / ADMIN_PASSWORD ortam degiskenleri ile)

import (
	"context"
	"log"
	"os"

	"seyahat/backend/internal/config"
	"seyahat/backend/internal/database"
	"seyahat/backend/internal/repository"
	"seyahat/backend/internal/service"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("veritabani baglantisi kurulamadi: %v", err)
	}
	defer pool.Close()

	email := os.Getenv("ADMIN_EMAIL")
	if email == "" {
		email = "admin@example.com"
	}
	pass := os.Getenv("ADMIN_PASSWORD")
	if pass == "" {
		pass = "Admin123!"
	}

	hash, err := service.HashPassword(pass)
	if err != nil {
		log.Fatalf("sifre hashlenemedi: %v", err)
	}

	repo := repository.NewUserRepository(pool)
	id, err := repo.UpsertAdmin(ctx, "Sistem Yoneticisi", email, hash)
	if err != nil {
		log.Fatalf("admin olusturulamadi: %v", err)
	}
	log.Printf("Admin hazir: %s (id=%s)", email, id)
}
