CREATE TABLE users (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	password_hash VARCHAR(255) NOT NULL,
	status ENUM('active', 'deleted') NOT NULL DEFAULT 'active',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	deleted_at TIMESTAMP NULL,
	CHECK (CHAR_LENGTH(TRIM(name)) > 0),
	CHECK (
		(
			status = 'active'
			AND deleted_at IS NULL
		)
		OR (
			status = 'deleted'
			AND deleted_at IS NOT NULL
		)
	)
);
CREATE TABLE teams (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	name VARCHAR(255) NOT NULL,
	created_by BIGINT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
	UNIQUE (created_by, name),
	CHECK (CHAR_LENGTH(TRIM(name)) > 0)
);
CREATE TABLE team_members (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	team_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,
	role ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	UNIQUE (team_id, user_id)
);
CREATE TABLE tasks (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	team_id BIGINT NOT NULL,
	title VARCHAR(255) NOT NULL,
	description TEXT NULL,
	status ENUM('todo', 'in_progress', 'done') NOT NULL DEFAULT 'todo',
	assignee_id BIGINT NULL,
	created_by BIGINT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	done_at TIMESTAMP NULL,
	FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
	FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
	FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
	CHECK (CHAR_LENGTH(TRIM(title)) > 0),
	CHECK (
		(
			status = 'done'
			AND done_at IS NOT NULL
		)
		OR (
			status <> 'done'
			AND done_at IS NULL
		)
	)
);
CREATE TABLE task_history (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	task_id BIGINT NOT NULL,
	changed_by BIGINT NOT NULL,
	action VARCHAR(50) NOT NULL,
	field_name VARCHAR(50) NULL,
	old_value TEXT NULL,
	new_value TEXT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE RESTRICT,
	CHECK (CHAR_LENGTH(TRIM(action)) > 0)
);
CREATE TABLE task_comments (
	id BIGINT PRIMARY KEY AUTO_INCREMENT,
	task_id BIGINT NOT NULL,
	user_id BIGINT NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
	CHECK (CHAR_LENGTH(TRIM(content)) > 0)
);
CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_tasks_team_assignee ON tasks(team_id, assignee_id);
CREATE INDEX idx_task_history_task_created_at ON task_history(task_id, created_at);
CREATE INDEX idx_tasks_team_status_done_at ON tasks(team_id, status, done_at);