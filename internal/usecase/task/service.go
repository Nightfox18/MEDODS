package task

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	now := s.now()
	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		ScheduledAt: normalized.ScheduledAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		Recurrence:  normalized.Recurrence,
	}

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	// Получаем текущую задачу, чтобы сохранить TemplateID и CreatedAt
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		ScheduledAt: normalized.ScheduledAt,
		CreatedAt:   existing.CreatedAt, // сохраняем исходное время создания
		UpdatedAt:   s.now(),
		Recurrence:  normalized.Recurrence,
		TemplateID:  existing.TemplateID, // сохраняем ссылку на шаблон, если это экземпляр
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

// GenerateTasksForDate создаёт экземпляры задач на указанную дату на основе шаблонов.
func (s *Service) GenerateTasksForDate(ctx context.Context, date time.Time) error {
	templates, err := s.repo.FindTemplates(ctx)
	if err != nil {
		return err
	}

	date = date.Truncate(24 * time.Hour)

	for _, tmpl := range templates {
		if taskdomain.ShouldGenerateOnDate(tmpl.Recurrence, date, tmpl.ScheduledAt) {
			instance := *tmpl // копируем поля
			instance.ID = 0   // сброс ID для автоинкремента
			instance.TemplateID = &tmpl.ID
			instance.Recurrence = nil
			instance.ScheduledAt = date
			instance.Status = taskdomain.StatusNew

			now := s.now()
			instance.CreatedAt = now
			instance.UpdatedAt = now

			if _, err := s.repo.Create(ctx, &instance); err != nil {
				// Логируем ошибку, но продолжаем генерацию других задач
				log.Printf("failed to create instance from template %d: %v", tmpl.ID, err)
			}
		}
	}
	return nil
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.ScheduledAt.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: scheduled_at is required", ErrInvalidInput)
	}
	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}
	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if input.Recurrence != nil {
		if err := input.Recurrence.Validate(); err != nil {
			return CreateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}
	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}
	if input.ScheduledAt.IsZero() {
		return UpdateInput{}, fmt.Errorf("%w: scheduled_at is required", ErrInvalidInput)
	}
	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	if input.Recurrence != nil {
		if err := input.Recurrence.Validate(); err != nil {
			return UpdateInput{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		}
	}
	return input, nil
}