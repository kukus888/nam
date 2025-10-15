-- Add SSH connection support fields to servers table
ALTER TABLE server 
ADD COLUMN ssh_port INTEGER DEFAULT 22,
ADD COLUMN ssh_auth_type VARCHAR(20) DEFAULT 'password', -- 'password' or 'private_key'
ADD COLUMN ssh_auth_secret_id BIGINT REFERENCES secrets(id) ON DELETE SET NULL,
ADD COLUMN ssh_user VARCHAR(255);

-- Add index for SSH secret lookups
CREATE INDEX IF NOT EXISTS idx_server_ssh_secret ON server(ssh_auth_secret_id);

-- Add comments for documentation
COMMENT ON COLUMN server.ssh_port IS 'SSH port to connect to on the server, default 22';
COMMENT ON COLUMN server.ssh_auth_type IS 'SSH authentication type: password or private_key';
COMMENT ON COLUMN server.ssh_auth_secret_id IS 'Foreign key reference to secrets table for SSH credentials';
COMMENT ON COLUMN server.ssh_user IS 'Username for SSH connection to the server';