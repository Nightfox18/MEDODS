package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id
	`

	var recurrenceJSON []byte
	if task.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(task.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.ScheduledAt,
		task.CreatedAt, task.UpdatedAt, recurrenceJSON, task.TemplateID,
	)

	return scanTask(row)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id
		FROM tasks
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	task, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}
	return task, nil
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1,
			description = $2,
			status = $3,
			scheduled_at = $4,
			updated_at = $5,
			recurrence = $6,
			template_id = $7
		WHERE id = $8
		RETURNING id, title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id
	`

	var recurrenceJSON []byte
	if task.Recurrence != nil {
		var err error
		recurrenceJSON, err = json.Marshal(task.Recurrence)
		if err != nil {
			return nil, err
		}
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.ScheduledAt,
		task.UpdatedAt, recurrenceJSON, task.TemplateID, task.ID,
	)

	updated, err := scanTask(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, taskdomain.ErrNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id
		FROM tasks
		ORDER BY scheduled_at DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]taskdomain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindTemplates возвращает все задачи, у которых заполнено recurrence (шаблоны).
func (r *Repository) FindTemplates(ctx context.Context) ([]*taskdomain.Task, error) {
	const query = `
		SELECT id, title, description, status, scheduled_at, created_at, updated_at, recurrence, template_id
		FROM tasks
		WHERE recurrence IS NOT NULL
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]*taskdomain.Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return templates, nil
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		recurrenceJSON []byte
	)

	err := scanner.Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&status,
		&task.ScheduledAt,
		&task.CreatedAt,
		&task.UpdatedAt,
		&recurrenceJSON,
		&task.TemplateID,
	)
	if err != nil {
		return nil, err
	}

	task.Status = taskdomain.Status(status)

	if len(recurrenceJSON) > 0 {
		var rec taskdomain.Recurrence
		if err := json.Unmarshal(recurrenceJSON, &rec); err != nil {
			return nil, err
		}
		task.Recurrence = &rec
	}

	return &task, nil
}
