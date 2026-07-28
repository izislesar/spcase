//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"spcase.ru/backend/internal/domain"
)

func TestUserPostgresCreatesFirstAdminOnlyOnce(t *testing.T) {
	resetIntegrationDatabase(t)
	users, err := NewUserPostgres(integrationPool)
	if err != nil {
		t.Fatal(err)
	}

	first, err := users.CreateFirstAdmin(context.Background(), domain.User{
		FullName:     "First Administrator",
		Email:        "first-admin@example.test",
		PasswordHash: "bcrypt-test-hash",
		Role:         domain.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("create first admin: %v", err)
	}
	if first.Role != domain.RoleAdmin || first.AuthVersion != 1 {
		t.Fatalf("unexpected first admin: %#v", first)
	}

	_, err = users.CreateFirstAdmin(context.Background(), domain.User{
		FullName:     "Second Administrator",
		Email:        "second-admin@example.test",
		PasswordHash: "another-bcrypt-test-hash",
		Role:         domain.RoleAdmin,
	})
	if !errors.Is(err, domain.ErrAdminAlreadyExists) {
		t.Fatalf("second bootstrap error = %v", err)
	}

	var adminCount int
	if err := integrationPool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM users WHERE role = 'ADMIN'`,
	).Scan(&adminCount); err != nil {
		t.Fatal(err)
	}
	if adminCount != 1 {
		t.Fatalf("admin count = %d", adminCount)
	}
}

func TestUserPostgresSerializesConcurrentAdminBootstrap(t *testing.T) {
	resetIntegrationDatabase(t)
	users, err := NewUserPostgres(integrationPool)
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for index := 0; index < 2; index++ {
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			<-start
			_, err := users.CreateFirstAdmin(context.Background(), domain.User{
				FullName:     fmt.Sprintf("Concurrent Administrator %d", index),
				Email:        fmt.Sprintf("concurrent-admin-%d@example.test", index),
				PasswordHash: "bcrypt-test-hash",
				Role:         domain.RoleAdmin,
			})
			results <- err
		}(index)
	}
	close(start)
	workers.Wait()
	close(results)

	var created, rejected int
	for err := range results {
		switch {
		case err == nil:
			created++
		case errors.Is(err, domain.ErrAdminAlreadyExists):
			rejected++
		default:
			t.Fatalf("unexpected concurrent bootstrap error: %v", err)
		}
	}
	if created != 1 || rejected != 1 {
		t.Fatalf("created = %d, rejected = %d", created, rejected)
	}
}
