-- Create issue_types table
CREATE TABLE issue_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50),
    color VARCHAR(7),
    is_subtask BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Insert default issue types
INSERT INTO issue_types (name, description, icon, color, is_subtask) VALUES
('Epic', 'Large feature or initiative', 'epic', '#6554C0', false),
('Story', 'User story or feature', 'story', '#0052CC', false),
('Task', 'Individual work item', 'task', '#2684FF', false),
('Bug', 'Defect or issue', 'bug', '#FF5630', false),
('Subtask', 'Sub-item of a parent issue', 'subtask', '#5E6C84', true);
