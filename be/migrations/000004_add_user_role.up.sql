-- 1. Drop old constraint if present from older schemas
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- 2. Allow legacy password_hash column to be nullable if present
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' AND column_name = 'password_hash'
    ) THEN
        ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
    END IF;
END $$;

-- 3. Add role column if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'viewer';

-- 4. Map existing legacy roles to valid kubenexus roles
UPDATE users SET role = 'admin' WHERE role IN ('super_admin', 'branch_manager') OR email LIKE '%admin%';
UPDATE users SET role = 'devops' WHERE role IN ('receptionist', 'housekeeper', 'maintenance');
UPDATE users SET role = 'viewer' WHERE role NOT IN ('admin', 'devops', 'viewer');

-- 5. Add check constraint for valid roles
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'devops', 'viewer'));

-- 6. Ensure default admin user admin@kubeenv.local exists
INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'Cluster Administrator',
    'admin@kubeenv.local',
    '$2a$10$4adB0n4lEzMEupT71Zle5.RasMuduUpko9blVBL4M1DCRLuifJM.a',
    'admin',
    true,
    NOW(),
    NOW()
) ON CONFLICT (email) DO UPDATE SET role = 'admin', password = EXCLUDED.password;

-- 7. Ensure index exists
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
