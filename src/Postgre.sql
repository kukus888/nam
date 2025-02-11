CREATE TABLE "TopologyNodes" (
  "ID" SERIAL SERIAL PRIMARY KEY,
  "name" VARCHAR,
  "type" VARCHAR
);

CREATE TABLE "Proxy" (
  "ID" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("ID"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("ID"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("ID")
);

CREATE TABLE "F5" (
  "ID" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("ID"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("ID")
);

CREATE TABLE "F5Egress" (
  "ID" SERIAL PRIMARY KEY REFERENCES "F5" ("ID"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("ID")
);

CREATE TABLE "Nginx" (
  "ID" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("ID"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("ID")
);

CREATE TABLE "NginxEgress" (
  "ID" SERIAL PRIMARY KEY REFERENCES "Nginx" ("ID"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("ID")
);

CREATE TABLE "ApplicationDefinitions" (
  "ID" SERIAL PRIMARY KEY,
  "name" VARCHAR,
  "port" integer,
  "type" VARCHAR
);

CREATE TABLE "Servers" (
  "ID" SERIAL PRIMARY KEY,
  "alias" VARCHAR,
  "hostname" VARCHAR UNIQUE
);

CREATE TABLE "ApplicationInstances" (
  "ID" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("ID"),
  "server" SERIAL REFERENCES "Servers" ("ID"),
  "definition" integer REFERENCES "ApplicationDefinitions" ("ID")
);

CREATE TABLE "Healthchecks" (
  "ID" SERIAL PRIMARY KEY,
  "application" SERIAL REFERENCES "ApplicationDefinitions" ("ID"),
  "url" VARCHAR,
  "timeout" time,
  "interval" time,
  "expectedstatus" int
);
