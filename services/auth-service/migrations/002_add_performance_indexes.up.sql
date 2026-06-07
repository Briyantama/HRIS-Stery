-- Add indexes for frequently used query patterns

-- auth-service: GetForRoles queries by role_id
-- This index supports the ANY() clause in permission lookups
CREATE INDEX idx_role_permissions_role_id ON auth.role_permissions(role_id);

-- auth-service: GetRoles queries user_roles by user_id (already covered by PRIMARY KEY)
-- No additional index needed - (user_id, role_id) PK covers it

-- Support fast lookups by role for cleaning up assignments
CREATE INDEX idx_user_roles_role_id ON auth.user_roles(role_id);
