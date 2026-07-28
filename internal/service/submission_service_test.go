package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"spcase.ru/backend/internal/domain"
)

type recordingSubmissionRepository struct {
	calls       int
	solutionURL string
	err         error
}

func (r *recordingSubmissionRepository) Upsert(
	_ context.Context,
	_ uuid.UUID,
	solutionURL string,
	_ time.Time,
) (domain.Submission, error) {
	r.calls++
	r.solutionURL = solutionURL
	if r.err != nil {
		return domain.Submission{}, r.err
	}
	return domain.Submission{SolutionURL: solutionURL}, nil
}

func TestSubmissionServicePersistsValidHTTPAndHTTPSURLs(t *testing.T) {
	t.Parallel()

	for _, rawURL := range []string{
		"http://example.com/result",
		" https://example.com/result ",
	} {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()

			repository := &recordingSubmissionRepository{}
			submissions := newSubmissionServiceForTest(t, repository)
			result, err := submissions.Submit(context.Background(), uuid.New(), rawURL)
			if err != nil {
				t.Fatalf("Submit(%q): %v", rawURL, err)
			}
			wantURL := strings.TrimSpace(rawURL)
			if repository.calls != 1 || repository.solutionURL != wantURL {
				t.Fatalf("repository calls = %d, URL = %q", repository.calls, repository.solutionURL)
			}
			if result.SolutionURL != wantURL {
				t.Fatalf("result URL = %q, want %q", result.SolutionURL, wantURL)
			}
		})
	}
}

func TestSubmissionServiceValidatesURLBeforePersistence(t *testing.T) {
	t.Parallel()

	repository := &recordingSubmissionRepository{}
	submissions := newSubmissionServiceForTest(t, repository)
	overlongURL := "https://example.com/" +
		strings.Repeat("a", domain.MaximumSolutionURLLength)

	for _, rawURL := range []string{
		"ftp://example.com/result",
		"https://exa mple.com/result",
		overlongURL,
	} {
		_, err := submissions.Submit(context.Background(), uuid.New(), rawURL)
		if !errors.Is(err, domain.ErrInvalidURLFormat) {
			t.Fatalf("Submit(%q) error = %v, want %v", rawURL, err, domain.ErrInvalidURLFormat)
		}
	}
	if repository.calls != 0 {
		t.Fatalf("repository calls = %d, want 0 for invalid URLs", repository.calls)
	}
}

func TestSubmissionServicePreservesSubmissionRules(t *testing.T) {
	t.Parallel()

	t.Run("deadline", func(t *testing.T) {
		repository := &recordingSubmissionRepository{}
		submissions := newSubmissionServiceForTest(t, repository)
		submissions.now = func() time.Time { return time.Now().Add(2 * time.Hour) }

		_, err := submissions.Submit(
			context.Background(),
			uuid.New(),
			"https://example.com/result",
		)
		if !errors.Is(err, domain.ErrDeadlinePassed) {
			t.Fatalf("Submit() error = %v, want %v", err, domain.ErrDeadlinePassed)
		}
		if repository.calls != 0 {
			t.Fatalf("repository calls = %d, want 0 after deadline", repository.calls)
		}
	})

	t.Run("team membership rule", func(t *testing.T) {
		repository := &recordingSubmissionRepository{err: domain.ErrMinimumTwoMembers}
		submissions := newSubmissionServiceForTest(t, repository)

		_, err := submissions.Submit(
			context.Background(),
			uuid.New(),
			"https://example.com/result",
		)
		if !errors.Is(err, domain.ErrMinimumTwoMembers) {
			t.Fatalf("Submit() error = %v, want %v", err, domain.ErrMinimumTwoMembers)
		}
		if repository.calls != 1 {
			t.Fatalf("repository calls = %d, want 1", repository.calls)
		}
	})
}

func newSubmissionServiceForTest(
	t *testing.T,
	repository SubmissionRepository,
) *SubmissionService {
	t.Helper()

	submissions, err := NewSubmissionService(repository, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("NewSubmissionService(): %v", err)
	}
	return submissions
}
