package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"

	"spcase.ru/backend/internal/domain"
)

type ExportRepository interface {
	ExportSummary(context.Context) ([]domain.ExportSummaryRow, error)
	ExportDetails(context.Context) ([]domain.ExportDetailRow, error)
}

type ExportService struct {
	repository ExportRepository
}

func NewExportService(repository ExportRepository) (*ExportService, error) {
	if repository == nil {
		return nil, errors.New("export repository cannot be nil")
	}
	return &ExportService{repository: repository}, nil
}

func (s *ExportService) WriteXLSX(ctx context.Context, writer io.Writer) error {
	summary, err := s.repository.ExportSummary(ctx)
	if err != nil {
		return err
	}
	details, err := s.repository.ExportDetails(ctx)
	if err != nil {
		return err
	}

	file := excelize.NewFile()
	defer file.Close()
	summarySheet := "Сводка"
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, summarySheet); err != nil {
		return fmt.Errorf("name summary sheet: %w", err)
	}
	detailSheet := "Детализация по жюри"
	if _, err := file.NewSheet(detailSheet); err != nil {
		return fmt.Errorf("create detail sheet: %w", err)
	}
	summaryHeader := []any{
		"Команда", "Капитан", "Telegram капитана", "Состав",
		"Участников", "Решение", "Итоговый балл", "Оценило жюри",
	}
	if err := file.SetSheetRow(summarySheet, "A1", &summaryHeader); err != nil {
		return err
	}
	for index, row := range summary {
		values := []any{
			row.TeamName, row.CaptainName, row.CaptainTelegram, row.Members,
			row.TotalMembers, row.SolutionURL, row.TotalScore, row.EvaluatedByCount,
		}
		cell := fmt.Sprintf("A%d", index+2)
		if err := file.SetSheetRow(summarySheet, cell, &values); err != nil {
			return fmt.Errorf("write summary row: %w", err)
		}
	}
	detailHeader := []any{"Команда", "Член жюри", "Критерий", "Оценка"}
	if err := file.SetSheetRow(detailSheet, "A1", &detailHeader); err != nil {
		return err
	}
	for index, row := range details {
		values := []any{row.TeamName, row.JuryName, row.CriterionID, row.Score}
		cell := fmt.Sprintf("A%d", index+2)
		if err := file.SetSheetRow(detailSheet, cell, &values); err != nil {
			return fmt.Errorf("write detail row: %w", err)
		}
	}
	if err := file.SetColWidth(summarySheet, "A", "H", 22); err != nil {
		return err
	}
	if err := file.SetColWidth(detailSheet, "A", "D", 22); err != nil {
		return err
	}
	if err := file.Write(writer); err != nil {
		return fmt.Errorf("stream XLSX: %w", err)
	}
	return nil
}
