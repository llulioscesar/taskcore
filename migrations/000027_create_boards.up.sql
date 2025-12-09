CREATE TABLE boards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'kanban',
    filter_jql TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_boards_project ON boards(project_id);

CREATE TABLE board_columns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    board_id UUID NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    status_id UUID REFERENCES workflow_statuses(id) ON DELETE SET NULL,
    position INTEGER NOT NULL DEFAULT 0,
    min_limit INTEGER,
    max_limit INTEGER
);

CREATE INDEX idx_board_columns_board ON board_columns(board_id);

CREATE TABLE board_configs (
    board_id UUID PRIMARY KEY REFERENCES boards(id) ON DELETE CASCADE,
    swimlane VARCHAR(20) NOT NULL DEFAULT 'none',
    card_fields TEXT[] DEFAULT '{}',
    show_days_in_col BOOLEAN NOT NULL DEFAULT FALSE,
    show_epic_as_bar BOOLEAN NOT NULL DEFAULT FALSE
);
