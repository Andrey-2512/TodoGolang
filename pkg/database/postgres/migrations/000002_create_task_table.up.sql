CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT,
    description TEXT,
    user_id BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE

);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);