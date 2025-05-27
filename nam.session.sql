DROP TABLE IF EXISTS application_instance CASCADE;

CREATE TABLE IF NOT EXISTS application_instance (
  id SERIAL PRIMARY KEY,
  topology_node_id SERIAL REFERENCES topology_node (id),
  name VARCHAR UNIQUE,
  server_id SERIAL REFERENCES server (id),
  application_definition_id SERIAL REFERENCES application_definition (id)
);

DROP TABLE IF EXISTS healthcheck_results CASCADE;

CREATE TABLE IF NOT EXISTS healthcheck_results (
	id BIGSERIAL,
	healthcheck_id SERIAL NOT NULL REFERENCES healthcheck (id),
  application_instance_id INTEGER NOT NULL REFERENCES application_instance (id),
	is_successful BOOLEAN NOT NULL,
	time_start TIMESTAMPTZ NOT NULL,
	time_end TIMESTAMPTZ NOT NULL,
	res_status INTEGER NOT NULL,
	res_body TEXT,
	res_time INTEGER NOT NULL, -- in milliseconds
	error_message TEXT
);