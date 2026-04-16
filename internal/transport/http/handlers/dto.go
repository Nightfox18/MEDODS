package handlers

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      taskdomain.Status     `json:"status"`
	ScheduledAt time.Time             `json:"scheduled_at"`
	Recurrence  *taskdomain.Recurrence `json:"recurrence,omitempty"`
}

type taskDTO struct {
	ID          int64                 `json:"id"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Status      taskdomain.Status     `json:"status"`
	ScheduledAt time.Time             `json:"scheduled_at"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Recurrence  *taskdomain.Recurrence `json:"recurrence,omitempty"`
	TemplateID  *int64                `json:"template_id,omitempty"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		ScheduledAt: task.ScheduledAt,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		Recurrence:  task.Recurrence,
		TemplateID:  task.TemplateID,
	}
}