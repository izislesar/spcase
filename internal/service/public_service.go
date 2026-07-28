package service

import (
	_ "embed"
	"encoding/json"
	"errors"
	"time"
)

//go:embed content/schedule.json
var scheduleJSON []byte

//go:embed content/faq.json
var faqJSON []byte

type ScheduleEvent struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	StartTime   time.Time `json:"start_time"`
	Description string    `json:"description"`
}

type FAQItem struct {
	ID       int    `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type PublicInfo struct {
	RegistrationDeadline time.Time
	SubmissionDeadline   time.Time
	RegistrationOpen     bool
	SubmissionOpen       bool
}

type PublicService struct {
	registrationDeadline time.Time
	submissionDeadline   time.Time
	noTeamTelegramURL    string
	schedule             []ScheduleEvent
	faq                  []FAQItem
	now                  func() time.Time
}

func NewPublicService(
	registrationDeadline, submissionDeadline time.Time,
	noTeamTelegramURL string,
) (*PublicService, error) {
	if registrationDeadline.IsZero() || submissionDeadline.IsZero() || noTeamTelegramURL == "" {
		return nil, errors.New("public configuration is incomplete")
	}
	var schedule []ScheduleEvent
	if err := json.Unmarshal(scheduleJSON, &schedule); err != nil {
		return nil, err
	}
	var faq []FAQItem
	if err := json.Unmarshal(faqJSON, &faq); err != nil {
		return nil, err
	}
	return &PublicService{
		registrationDeadline: registrationDeadline.UTC(),
		submissionDeadline:   submissionDeadline.UTC(),
		noTeamTelegramURL:    noTeamTelegramURL,
		schedule:             schedule,
		faq:                  faq,
		now:                  time.Now,
	}, nil
}

func (s *PublicService) Info() PublicInfo {
	now := s.now().UTC()
	return PublicInfo{
		RegistrationDeadline: s.registrationDeadline,
		SubmissionDeadline:   s.submissionDeadline,
		RegistrationOpen:     now.Before(s.registrationDeadline),
		SubmissionOpen:       now.Before(s.submissionDeadline),
	}
}

func (s *PublicService) Schedule() []ScheduleEvent {
	return append([]ScheduleEvent(nil), s.schedule...)
}

func (s *PublicService) FAQ() []FAQItem {
	return append([]FAQItem(nil), s.faq...)
}

func (s *PublicService) NoTeam() (string, string) {
	return "Если у вас нет команды, перейдите в закрытый Telegram-чат для поиска сокомандников.", s.noTeamTelegramURL
}
