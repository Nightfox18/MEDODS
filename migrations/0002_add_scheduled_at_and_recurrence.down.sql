DROP INDEX IF EXISTS idx_tasks_recurrence_not_null;
DROP INDEX IF EXISTS idx_tasks_template_id;
ALTER TABLE tasks DROP COLUMN template_id;
ALTER TABLE tasks DROP COLUMN recurrence;
ALTER TABLE tasks DROP COLUMN scheduled_at;