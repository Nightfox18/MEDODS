package task

import (
	"errors"
	"fmt"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

type RecurrenceType string

const (
	RecurrenceNone     RecurrenceType = ""
	RecurrenceDaily    RecurrenceType = "daily"
	RecurrenceMonthly  RecurrenceType = "monthly"
	RecurrenceSpecific RecurrenceType = "specific"
	RecurrenceParity   RecurrenceType = "parity"
)

// Recurrence описывает настройки повторения задачи.
type Recurrence struct {
	Type       RecurrenceType `json:"type"`
	Interval   int            `json:"interval,omitempty"`    // для daily
	DayOfMonth int            `json:"day_of_month,omitempty"` // для monthly (1-30)
	Dates      []string       `json:"dates,omitempty"`       // для specific (формат "2006-01-02")
	Parity     string         `json:"parity,omitempty"`      // "even" или "odd"
}

// Validate проверяет корректность настроек.
func (r *Recurrence) Validate() error {
	if r == nil {
		return nil
	}
	switch r.Type {
	case RecurrenceDaily:
		if r.Interval <= 0 {
			return errors.New("interval must be positive")
		}
	case RecurrenceMonthly:
		if r.DayOfMonth < 1 || r.DayOfMonth > 30 {
			return errors.New("day_of_month must be between 1 and 30")
		}
	case RecurrenceSpecific:
		if len(r.Dates) == 0 {
			return errors.New("at least one date required")
		}
		for _, d := range r.Dates {
			if _, err := time.Parse("2006-01-02", d); err != nil {
				return fmt.Errorf("invalid date format: %s", d)
			}
		}
	case RecurrenceParity:
		if r.Parity != "even" && r.Parity != "odd" {
			return errors.New("parity must be 'even' or 'odd'")
		}
	default:
		return errors.New("unknown recurrence type")
	}
	return nil
}

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	ScheduledAt time.Time   `json:"scheduled_at"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Recurrence  *Recurrence `json:"recurrence,omitempty"`
	TemplateID  *int64      `json:"template_id,omitempty"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusNew, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}