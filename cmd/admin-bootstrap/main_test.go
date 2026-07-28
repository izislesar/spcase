package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"spcase.ru/backend/internal/domain"
	"spcase.ru/backend/internal/service"
)

type recordingBootstrapper struct {
	input service.AdminBootstrapInput
}

func (b *recordingBootstrapper) Bootstrap(
	_ context.Context,
	input service.AdminBootstrapInput,
) (domain.User, error) {
	b.input = input
	return domain.User{Role: domain.RoleAdmin}, nil
}

func TestExecuteDoesNotExposeSensitiveInput(t *testing.T) {
	const (
		fullName = "Sensitive Administrator Name"
		email    = "sensitive-admin@example.com"
		password = "sensitive-bootstrap-password-42"
	)
	bootstrap := &recordingBootstrapper{}
	var stdout, stderr bytes.Buffer

	err := execute(
		context.Background(),
		bootstrap,
		[]string{"-full-name", fullName, "-email", email},
		strings.NewReader(password+"\n"),
		&stdout,
		&stderr,
	)
	if err != nil {
		t.Fatal(err)
	}
	if bootstrap.input.FullName != fullName ||
		bootstrap.input.Email != email ||
		bootstrap.input.Password != password {
		t.Fatalf("bootstrap input was not passed intact: %#v", bootstrap.input)
	}

	output := stdout.String() + stderr.String()
	for _, sensitive := range []string{fullName, email, password} {
		if strings.Contains(output, sensitive) {
			t.Fatalf("sensitive value %q was printed in %q", sensitive, output)
		}
	}
	if strings.TrimSpace(stdout.String()) != "administrator bootstrap completed" {
		t.Fatalf("unexpected stdout: %q", stdout.String())
	}
}

func TestReadPasswordRejectsMultipleLines(t *testing.T) {
	if _, err := readPassword(strings.NewReader("first\nsecond\n")); err == nil {
		t.Fatal("multiple password lines were accepted")
	}
}
