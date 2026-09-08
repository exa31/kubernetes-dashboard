-- Add allowed_namespaces column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS allowed_namespaces VARCHAR(500) NOT NULL DEFAULT '*';

-- Update existing users to have default '*'
UPDATE users SET allowed_namespaces = '*' WHERE allowed_namespaces IS NULL OR allowed_namespaces = '';
