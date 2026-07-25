package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/otelforge/otelforge/internal/config"
	"github.com/otelforge/otelforge/internal/crypto"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/models"
)

func main() {
	email := flag.String("email", "", "user email")
	password := flag.String("password", "", "user password")
	role := flag.String("role", "admin", "user role: user or admin")
	flag.Parse()

	if *email == "" || *password == "" {
		fmt.Fprintln(os.Stderr, "usage: seed --email x --password y [--role admin]")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}
	store := db.NewStore(pool)
	hash, err := crypto.HashPassword(*password)
	if err != nil {
		log.Fatal(err)
	}
	user, err := store.CreateUser(ctx, *email, hash, models.Role(*role))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("created user %s (%s) role=%s", user.Email, user.ID, user.Role)
}
