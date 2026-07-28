package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"spcase.ru/backend/internal/config"
	"spcase.ru/backend/internal/domain"
	postgrespool "spcase.ru/backend/internal/pkg/postgres"
	"spcase.ru/backend/internal/repository"
	"spcase.ru/backend/internal/service"
)

const (
	bootstrapTimeout          = 30 * time.Second
	maximumPasswordInputBytes = 4096
)

type adminBootstrapper interface {
	Bootstrap(context.Context, service.AdminBootstrapInput) (domain.User, error)
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "administrator bootstrap failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	ctx, cancel := context.WithTimeout(ctx, bootstrapTimeout)
	defer cancel()

	pool, err := postgrespool.New(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()

	users, err := repository.NewUserPostgres(pool)
	if err != nil {
		return err
	}
	bootstrap, err := service.NewAdminBootstrapService(users)
	if err != nil {
		return err
	}
	return execute(ctx, bootstrap, args, stdin, stdout, stderr)
}

func execute(
	ctx context.Context,
	bootstrap adminBootstrapper,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) error {
	if bootstrap == nil {
		return errors.New("admin bootstrap service cannot be nil")
	}

	flags := flag.NewFlagSet("admin-bootstrap", flag.ContinueOnError)
	flags.SetOutput(stderr)
	fullName := flags.String("full-name", "", "administrator's full name")
	email := flags.String("email", "", "administrator's email address")
	flags.Usage = func() {
		_, _ = fmt.Fprintln(stderr,
			"usage: admin-bootstrap -full-name <name> -email <email> < password-file")
	}
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		flags.Usage()
		return errors.New("positional arguments are not accepted")
	}

	password, err := readPassword(stdin)
	if err != nil {
		return err
	}
	_, err = bootstrap.Bootstrap(ctx, service.AdminBootstrapInput{
		FullName: *fullName,
		Email:    *email,
		Password: password,
	})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdout, "administrator bootstrap completed")
	return err
}

func readPassword(reader io.Reader) (string, error) {
	if reader == nil {
		return "", errors.New("password input cannot be nil")
	}
	input, err := io.ReadAll(io.LimitReader(reader, maximumPasswordInputBytes+1))
	if err != nil {
		return "", errors.New("read password from stdin")
	}
	if len(input) > maximumPasswordInputBytes {
		return "", errors.New("password input is too large")
	}
	input = bytes.TrimSuffix(input, []byte("\n"))
	input = bytes.TrimSuffix(input, []byte("\r"))
	if bytes.ContainsAny(input, "\r\n") {
		return "", errors.New("password input must contain exactly one line")
	}
	if len(input) == 0 {
		return "", errors.New("password must be provided through stdin")
	}
	return string(input), nil
}
