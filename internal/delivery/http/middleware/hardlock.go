package middleware

import (
	"errors"
	"net/http"

	"spcase.ru/backend/internal/domain"
)

// MutationLock reports the current team-mutation lock state.
type MutationLock interface {
	MutationsLocked() bool
}

// HardLockMiddleware blocks team composition changes during the final hour
// before the submission deadline.
type HardLockMiddleware struct {
	lock MutationLock
}

// NewHardLockMiddleware creates team-mutation hard-lock middleware.
func NewHardLockMiddleware(lock MutationLock) (*HardLockMiddleware, error) {
	if lock == nil {
		return nil, errors.New("mutation lock cannot be nil")
	}
	return &HardLockMiddleware{lock: lock}, nil
}

// Middleware rejects a wrapped team-mutation route when the hard lock is
// active. It must wrap leave, kick, transfer-ownership, and disband routes;
// create, join, and submit routes must not use this middleware.
func (m *HardLockMiddleware) Middleware(next http.Handler) http.Handler {
	if next == nil {
		panic("hard-lock middleware next handler cannot be nil")
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if m.lock.MutationsLocked() {
			writeDomainError(writer, http.StatusForbidden, domain.ErrMutationsLocked)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
