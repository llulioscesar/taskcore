-- Create permission_schemes table
CREATE TABLE permission_schemes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Trigger for updated_at
CREATE TRIGGER update_permission_schemes_updated_at
    BEFORE UPDATE ON permission_schemes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert default permission scheme
INSERT INTO permission_schemes (name, description) VALUES
('Default Permission Scheme', 'Default permissions for new projects');
