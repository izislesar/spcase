package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"spcase.ru/backend/internal/config"
	httpmiddleware "spcase.ru/backend/internal/delivery/http/middleware"
	v1 "spcase.ru/backend/internal/delivery/http/v1"
	"spcase.ru/backend/internal/domain"
	postgrespool "spcase.ru/backend/internal/pkg/postgres"
	"spcase.ru/backend/internal/repository"
	"spcase.ru/backend/internal/service"
)

const (
	startupTimeout     = 10 * time.Second
	shutdownTimeout    = 15 * time.Second
	readHeaderTimeout  = 5 * time.Second
	readTimeout        = 15 * time.Second
	writeTimeout       = 30 * time.Second
	idleTimeout        = 60 * time.Second
	maximumHeaderBytes = 1 << 20
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("application stopped", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	signalCtx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	startupCtx, cancelStartup := context.WithTimeout(signalCtx, startupTimeout)
	defer cancelStartup()
	pool, err := postgrespool.New(startupCtx, cfg.DB)
	if err != nil {
		logger.Error(
			"database startup failed",
			slog.String("event", "database_startup_failed"),
			slog.String("error", err.Error()),
		)
		return err
	}
	defer func() {
		pool.Close()
		logger.Info(
			"database connection pool closed",
			slog.String("event", "database_pool_closed"),
		)
	}()

	handler, err := buildHandler(cfg, pool, logger)
	if err != nil {
		return fmt.Errorf("build HTTP handler: %w", err)
	}

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maximumHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("HTTP server started", slog.String("event", "http_server_started"), slog.Int("port", cfg.Port))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-signalCtx.Done():
		logger.Info("graceful shutdown started", slog.String("event", "graceful_shutdown_started"))
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		logger.Error(
			"graceful shutdown failed",
			slog.String("event", "graceful_shutdown_failed"),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	logger.Info("graceful shutdown completed", slog.String("event", "graceful_shutdown_completed"))
	return nil
}

func buildHandler(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) (http.Handler, error) {
	userRepository, err := repository.NewUserPostgres(pool)
	if err != nil {
		return nil, err
	}
	teamRepository, err := repository.NewTeamPostgres(pool)
	if err != nil {
		return nil, err
	}
	scoreRepository, err := repository.NewScorePostgres(pool)
	if err != nil {
		return nil, err
	}
	submissionRepository, err := repository.NewSubmissionPostgres(pool)
	if err != nil {
		return nil, err
	}
	queryRepository, err := repository.NewQueryPostgres(pool)
	if err != nil {
		return nil, err
	}

	authService, err := service.NewAuthService(
		userRepository,
		cfg.JWT.Secret,
		cfg.JuryRegistrationKey,
		cfg.RegistrationDeadline,
	)
	if err != nil {
		return nil, err
	}
	teamService, err := service.NewTeamService(
		teamRepository, submissionRepository, cfg.SubmissionDeadline,
	)
	if err != nil {
		return nil, err
	}
	submissionService, err := service.NewSubmissionService(submissionRepository, cfg.SubmissionDeadline)
	if err != nil {
		return nil, err
	}
	scoreService, err := service.NewScoreService(scoreRepository)
	if err != nil {
		return nil, err
	}
	juryService, err := service.NewJuryService(queryRepository)
	if err != nil {
		return nil, err
	}
	userService, err := service.NewUserService(userRepository, teamRepository)
	if err != nil {
		return nil, err
	}
	publicService, err := service.NewPublicService(
		cfg.RegistrationDeadline, cfg.SubmissionDeadline, cfg.NoTeamTelegramURL,
	)
	if err != nil {
		return nil, err
	}
	adminService, err := service.NewAdminService(queryRepository)
	if err != nil {
		return nil, err
	}
	exportService, err := service.NewExportService(queryRepository)
	if err != nil {
		return nil, err
	}

	authHandler, err := v1.NewAuthHandler(authService, cfg.AppDomain, logger)
	if err != nil {
		return nil, err
	}
	userHandler, err := v1.NewUserHandler(userService, logger)
	if err != nil {
		return nil, err
	}
	teamHandler, err := v1.NewTeamHandler(teamService, logger)
	if err != nil {
		return nil, err
	}
	submissionHandler, err := v1.NewSubmissionHandler(submissionService, logger)
	if err != nil {
		return nil, err
	}
	juryHandler, err := v1.NewJuryHandler(juryService, logger)
	if err != nil {
		return nil, err
	}
	scoreHandler, err := v1.NewScoreHandler(scoreService, logger)
	if err != nil {
		return nil, err
	}
	publicHandler, err := v1.NewPublicHandler(publicService, pool, logger)
	if err != nil {
		return nil, err
	}
	adminHandler, err := v1.NewAdminHandler(adminService, exportService, logger)
	if err != nil {
		return nil, err
	}

	authMiddleware, err := httpmiddleware.NewAuthMiddleware(authService, logger)
	if err != nil {
		return nil, err
	}
	hardLockMiddleware, err := httpmiddleware.NewHardLockMiddleware(teamService)
	if err != nil {
		return nil, err
	}
	corsMiddleware, err := httpmiddleware.NewCORSMiddleware(cfg.CORSAllowedOrigins)
	if err != nil {
		return nil, err
	}
	recoveryMiddleware := httpmiddleware.NewRecoveryMiddleware(logger)
	requestLoggingMiddleware := httpmiddleware.NewRequestLogging(logger)

	protected := func(handler http.HandlerFunc, roles ...domain.Role) http.Handler {
		return authMiddleware.Middleware(httpmiddleware.RequireRoles(roles...)(handler))
	}
	protectedMutation := func(handler http.HandlerFunc) http.Handler {
		return authMiddleware.Middleware(
			httpmiddleware.RequireRoles(domain.RoleUser)(hardLockMiddleware.Middleware(handler)),
		)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health/live", publicHandler.Live)
	mux.HandleFunc("GET /api/v1/health/ready", publicHandler.Ready)
	mux.HandleFunc("GET /api/v1/info", publicHandler.Info)
	mux.HandleFunc("GET /api/v1/schedule", publicHandler.Schedule)
	mux.HandleFunc("GET /api/v1/faq", publicHandler.FAQ)
	mux.HandleFunc("GET /api/v1/no-team", publicHandler.NoTeam)
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/jury/register", authHandler.RegisterJury)
	mux.HandleFunc("POST /api/v1/jury/login", authHandler.LoginJury)
	mux.Handle("POST /api/v1/auth/logout", protected(
		authHandler.Logout, domain.RoleUser, domain.RoleJury, domain.RoleAdmin,
	))
	mux.Handle("GET /api/v1/user/me", protected(userHandler.Me, domain.RoleUser, domain.RoleAdmin))

	mux.Handle("POST /api/v1/team/create", protected(teamHandler.Create, domain.RoleUser))
	mux.Handle("POST /api/v1/team/join", protected(teamHandler.Join, domain.RoleUser))
	mux.Handle("GET /api/v1/team/my", protected(teamHandler.My, domain.RoleUser))
	mux.Handle("POST /api/v1/team/leave", protectedMutation(teamHandler.Leave))
	mux.Handle("POST /api/v1/team/kick", protectedMutation(teamHandler.Kick))
	mux.Handle("POST /api/v1/team/transfer-ownership", protectedMutation(teamHandler.TransferOwnership))
	mux.Handle("DELETE /api/v1/team/disband", protectedMutation(teamHandler.Disband))
	mux.Handle("POST /api/v1/team/submit", protected(submissionHandler.Submit, domain.RoleUser))

	mux.Handle("GET /api/v1/jury/teams", protected(juryHandler.Teams, domain.RoleJury))
	mux.Handle("GET /api/v1/jury/evaluations", protected(scoreHandler.Evaluations, domain.RoleJury))
	mux.Handle("POST /api/v1/jury/evaluations", protected(scoreHandler.SaveEvaluations, domain.RoleJury))

	mux.Handle("GET /api/v1/admin/stats", protected(adminHandler.Stats, domain.RoleAdmin))
	mux.Handle("GET /api/v1/admin/export/excel", protected(adminHandler.ExportExcel, domain.RoleAdmin))
	mux.Handle("POST /api/v1/admin/evaluations/close", protected(
		adminHandler.CloseEvaluations, domain.RoleAdmin,
	))
	mux.Handle("POST /api/v1/admin/evaluations/open", protected(
		adminHandler.OpenEvaluations, domain.RoleAdmin,
	))

	return httpmiddleware.SecurityHeaders(
		httpmiddleware.NoStoreSensitiveResponses(
			httpmiddleware.RequestID(
				requestLoggingMiddleware.Middleware(
					corsMiddleware.Middleware(
						recoveryMiddleware.Middleware(httpmiddleware.APIErrorResponses(mux)),
					),
				),
			),
		),
	), nil
}
