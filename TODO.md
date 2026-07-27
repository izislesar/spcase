# Backend Implementation Roadmap

## Phase 1: Environment & Core
- [x] 1.1 Create `go.mod` and directory structure according to `backend.md`
- [x] 1.2 Implement `internal/config/config.go` (ENV loading with validation)
- [x] 1.3 Implement `internal/pkg/postgres/postgres.go` (`pgxpool` initialization, ping, timeouts)

## Phase 2: Domain Layer
- [x] 2.1 Implement `internal/domain/user.go` (Enums, structs)
- [x] 2.2 Implement `internal/domain/team.go` (Structs for team and members)
- [x] 2.3 Implement `internal/domain/score.go` (Structs for evaluation)
- [x] 2.4 Implement `internal/domain/errors.go` (Custom domain errors)
- [x] 2.5 Implement `internal/delivery/http/v1/dto.go` (Request/Response DTOs)

## Phase 3: Repository Layer (Database Access)
- [x] 3.1 Implement `internal/repository/user_postgres.go` (CRUD, auth queries)
- [x] 3.2 Implement `internal/repository/team_postgres.go` (Transactions with `FOR UPDATE` for Join/Kick/Leave)
- [x] 3.3 Implement `internal/repository/score_postgres.go` (Upsert jury scores, aggregation)

## Phase 4: Service Layer (Business Logic)
- [ ] 4.1 Implement `internal/service/auth_service.go` (Password hashing, JWT generation)
- [ ] 4.2 Implement `internal/service/team_service.go` (State machine, Hard Lock validation, invite logic)
- [ ] 4.3 Implement `internal/service/score_service.go` (Jury logic, calculation formulas)

## Phase 5: Delivery Layer (HTTP & Middlewares)
- [ ] 5.1 Implement `internal/delivery/http/middleware/auth.go` (JWT extraction from HttpOnly Cookie)
- [ ] 5.2 Implement `internal/delivery/http/middleware/hardlock.go` (Lock team mutations 1 hour before deadline)
- [ ] 5.3 Implement Handlers: `auth_handler.go`, `team_handler.go`, `score_handler.go`
- [ ] 5.4 Implement `main.go` (Dependency injection, HTTP server start, graceful shutdown)
