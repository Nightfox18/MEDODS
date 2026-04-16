-- Добавляем scheduled_at (обязательное поле, по умолчанию заполним текущим временем для существующих задач)
ALTER TABLE tasks ADD COLUMN scheduled_at TIMESTAMPTZ;
UPDATE tasks SET scheduled_at = created_at WHERE scheduled_at IS NULL;
ALTER TABLE tasks ALTER COLUMN scheduled_at SET NOT NULL;

-- Добавляем колонки для повторяющихся задач
ALTER TABLE tasks ADD COLUMN recurrence JSONB DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN template_id BIGINT DEFAULT NULL;

-- Индексы для эффективного поиска шаблонов и связи экземпляр-шаблон
CREATE INDEX idx_tasks_template_id ON tasks(template_id);
CREATE INDEX idx_tasks_recurrence_not_null ON tasks((1)) WHERE recurrence IS NOT NULL;