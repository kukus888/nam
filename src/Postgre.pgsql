CREATE TABLE IF NOT EXISTS "TopologyNode" (
  "ID" SERIAL PRIMARY KEY,
  "Name" VARCHAR,
  "Type" VARCHAR
);

CREATE TABLE IF NOT EXISTS "Proxy" (
  "ID" SERIAL REFERENCES "TopologyNode" ("ID"),
  "Ingress" SERIAL REFERENCES "TopologyNode" ("ID"),
  "Egress" SERIAL REFERENCES "TopologyNode" ("ID"),
  PRIMARY KEY ("ID")
);

CREATE TABLE IF NOT EXISTS "F5" (
  "ID" SERIAL REFERENCES "TopologyNode" ("ID"),
  "Ingress" SERIAL REFERENCES "TopologyNode" ("ID"),
  PRIMARY KEY ("ID")
);

CREATE TABLE IF NOT EXISTS "F5Egress" (
  "ID" SERIAL REFERENCES "F5" ("ID"),
  "Egress" SERIAL REFERENCES "TopologyNode" ("ID")
);

CREATE TABLE IF NOT EXISTS "Nginx" (
  "ID" SERIAL REFERENCES "TopologyNode" ("ID"),
  "Ingress" SERIAL REFERENCES "TopologyNode" ("ID"),
  PRIMARY KEY ("ID")
);

CREATE TABLE IF NOT EXISTS "NginxEgress" (
  "ID" SERIAL REFERENCES "Nginx" ("ID"),
  "Egress" SERIAL REFERENCES "TopologyNode" ("ID"),
  PRIMARY KEY ("ID")
);

CREATE TABLE IF NOT EXISTS "Healthcheck" (
  "ID" SERIAL PRIMARY KEY,
  "Url" VARCHAR,
  "Timeout" interval,
  "Interval" interval,
  "ExpectedStatus" int
);

CREATE TABLE IF NOT EXISTS "ApplicationDefinition" (
  "ID" SERIAL PRIMARY KEY,
  "HealthcheckId" SERIAL REFERENCES "Healthcheck" ("ID"),
  "Name" VARCHAR,
  "Port" integer,
  "Type" VARCHAR
);

CREATE TABLE IF NOT EXISTS "Server" (
  "ID" SERIAL PRIMARY KEY,
  "Alias" VARCHAR,
  "Hostname" VARCHAR UNIQUE
);

CREATE TABLE IF NOT EXISTS "ApplicationInstance" (
  "ID" SERIAL REFERENCES "TopologyNode" ("ID"),
  "ServerId" SERIAL REFERENCES "Server" ("ID"),
  "Definition" SERIAL REFERENCES "ApplicationDefinition" ("ID")
);


