-- Reverse: Remove performance indexes

DROP INDEX IF EXISTS auth.idx_role_permissions_role_id;
DROP INDEX IF EXISTS auth.idx_user_roles_role_id;
