CREATE TABLE "TopologyNodes" (
  "id" SERIAL SERIAL PRIMARY KEY,
  "name" VARCHAR,
  "type" VARCHAR
);

CREATE TABLE "Proxy" (
  "id" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("id"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("id"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("id")
);

CREATE TABLE "F5" (
  "id" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("id"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("id")
);

CREATE TABLE "F5Egress" (
  "id" SERIAL PRIMARY KEY REFERENCES "F5" ("id"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("id")
);

CREATE TABLE "Nginx" (
  "id" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("id"),
  "ingress" SERIAL REFERENCES "TopologyNodes" ("id")
);

CREATE TABLE "NginxEgress" (
  "id" SERIAL PRIMARY KEY REFERENCES "Nginx" ("id"),
  "egress" SERIAL REFERENCES "TopologyNodes" ("id")
);

CREATE TABLE "ApplicationDefinitions" (
  "id" SERIAL PRIMARY KEY,
  "name" VARCHAR,
  "port" integer,
  "type" VARCHAR
);

CREATE TABLE "Servers" (
  "id" SERIAL PRIMARY KEY,
  "alias" VARCHAR,
  "hostname" VARCHAR UNIQUE
);

CREATE TABLE "ApplicationInstances" (
  "id" SERIAL PRIMARY KEY REFERENCES "TopologyNodes" ("id"),
  "server" SERIAL REFERENCES "Servers" ("id"),
  "definition" integer REFERENCES "ApplicationDefinitions" ("id")
);

CREATE TABLE "Healthchecks" (
  "id" SERIAL PRIMARY KEY,
  "application" SERIAL REFERENCES "ApplicationDefinitions" ("id"),
  "url" VARCHAR,
  "timeout" time,
  "interval" time,
  "expectedstatus" int
);
