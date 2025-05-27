CREATE TABLE IF NOT EXISTS topology_node (
  id SERIAL PRIMARY KEY,
  type VARCHAR
);

CREATE TABLE IF NOT EXISTS proxy (
  id SERIAL,
  topology_node_id SERIAL REFERENCES topology_node (id),
  ingress SERIAL REFERENCES topology_node (id),
  egress SERIAL REFERENCES topology_node (id),
  name VARCHAR UNIQUE,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS f5 (
  id SERIAL,
  topology_node_id SERIAL REFERENCES topology_node (id),
  ingress SERIAL REFERENCES topology_node (id),
  name VARCHAR UNIQUE,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS f5_egress (
  id SERIAL REFERENCES f5 (id),
  egress SERIAL REFERENCES topology_node (id)
);

CREATE TABLE IF NOT EXISTS nginx (
  id SERIAL,
  topology_node_id SERIAL REFERENCES topology_node (id),
  name VARCHAR UNIQUE,
  ingress SERIAL REFERENCES topology_node (id),
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS nginx_egress (
  id SERIAL REFERENCES nginx (id),
  egress SERIAL REFERENCES topology_node (id),
  PRIMARY KEY (id)
);

-- CREATE TABLE statement for Healthcheck
CREATE TABLE healthcheck (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(2048) NOT NULL,
    method VARCHAR(10) DEFAULT 'GET',
    headers JSONB DEFAULT '[]',
    body TEXT,
    timeout BIGINT NOT NULL DEFAULT 5000000000, -- 5 seconds in nanoseconds
    check_interval BIGINT NOT NULL DEFAULT 60000000000, -- 60 seconds in nanoseconds
    retry_count INTEGER DEFAULT 3,
    retry_interval BIGINT DEFAULT 10000000000, -- 10 seconds in nanoseconds
    expected_status INTEGER DEFAULT 200,
    expected_response_body TEXT,
    response_validation VARCHAR(20) DEFAULT 'none', -- 'none', 'contains', 'exact', 'regex'
    
    verify_ssl BOOLEAN DEFAULT true,
    
    auth_type VARCHAR(20) DEFAULT 'none',
    auth_credentials TEXT
);

CREATE OR REPLACE FUNCTION notify_healthcheck_change()
RETURNS trigger AS $$
BEGIN
  PERFORM pg_notify('healthcheck_changes', TG_OP || ':' || COALESCE(NEW.id::text, OLD.id::text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS healthcheck_change_trigger ON healthcheck;
CREATE TRIGGER healthcheck_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON healthcheck
FOR EACH ROW
EXECUTE FUNCTION notify_healthcheck_change();

CREATE TABLE IF NOT EXISTS application_definition (
  id SERIAL PRIMARY KEY,
  healthcheck_id INTEGER REFERENCES healthcheck (id) NULL,
  name VARCHAR,
  port integer,
  type VARCHAR
);

CREATE TABLE IF NOT EXISTS server (
  id SERIAL PRIMARY KEY,
  alias VARCHAR,
  hostname VARCHAR UNIQUE
);

CREATE TABLE IF NOT EXISTS application_instance (
  id SERIAL,
  topology_node_id SERIAL REFERENCES topology_node (id),
  name VARCHAR UNIQUE,
  server_id SERIAL REFERENCES server (id),
  application_definition_id SERIAL REFERENCES application_definition (id)
);


