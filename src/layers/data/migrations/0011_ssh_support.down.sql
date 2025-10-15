-- Remove SSH connection support fields from servers table
ALTER TABLE server 
DROP COLUMN IF EXISTS ssh_port,
DROP COLUMN IF EXISTS ssh_auth_type,
DROP COLUMN IF EXISTS ssh_auth_secret_id,
DROP COLUMN IF EXISTS ssh_user;

-- Drop index
DROP INDEX IF EXISTS idx_server_ssh_secret;