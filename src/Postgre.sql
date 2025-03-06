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

CREATE TABLE IF NOT EXISTS healthcheck (
  id SERIAL PRIMARY KEY,
  name VARCHAR,
  url VARCHAR,
  timeout interval,
  check_interval interval,
  expected_status int
);

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


