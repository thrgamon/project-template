// Command provision-user creates an application user and an explicit Auth0
// identity mapping. It is an owner-operated bootstrap tool, not a signup API.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	sharedauth "github.com/thrgamon/infra/go/auth"

	"github.com/thrgamon/project-template/internal/config"
	"github.com/thrgamon/project-template/internal/db"
)

func main() {
	issuer := flag.String("issuer", "", "exact Auth0 issuer URL")
	subject := flag.String("subject", "", "exact Auth0 subject claim")
	email := flag.String("email", "", "email display attribute for the local user")
	name := flag.String("name", "", "display name")
	role := flag.String("role", "member", "application role")
	flag.Parse()
	if *issuer == "" || *subject == "" || *email == "" || *role == "" {
		fmt.Fprintln(os.Stderr, "usage: provision-user -issuer URL -subject SUBJECT -email EMAIL [-name NAME] [-role ROLE]")
		os.Exit(2)
	}

	ctx := context.Background()
	cfg := config.LoadConfig()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	queries := db.New(pool)
	user, err := queries.CreateUser(ctx, *email)
	if err != nil {
		log.Fatalf("create local user: %v", err)
	}
	store := sharedauth.NewPGStore(pool)
	claims := sharedauth.Claims{Issuer: *issuer, Subject: *subject, Email: *email, Name: *name}
	if err := store.ProvisionIdentity(ctx, claims, fmt.Sprint(user.ID), *role); err != nil {
		// An unmapped local user cannot authenticate. Try to leave no partial
		// bootstrap behind; report a failed compensation to the owner.
		if deleteErr := queries.DeleteUserByID(ctx, user.ID); deleteErr != nil {
			log.Printf("remove unlinked local user %d: %v", user.ID, deleteErr)
		}
		log.Fatalf("provision Auth0 identity: %v", err)
	}
	fmt.Printf("provisioned local user %d for %s %s with role %s\n", user.ID, *issuer, *subject, *role)
}
