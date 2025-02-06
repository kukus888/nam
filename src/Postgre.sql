CREATE TABLE "TopologyNodes" (
  "id" integer PRIMARY KEY,
  "name" string,
  "type" NetworkElementType
);

CREATE TABLE "Proxy" (
  "id" integer PRIMARY KEY,
  "ingress" integer,
  "egress" integer
);

CREATE TABLE "F5" (
  "id" integer,
  "ingress" integer,
  PRIMARY KEY ("id", "ingress")
);

CREATE TABLE "F5Egress" (
  "id" integer,
  "egress" integer,
  PRIMARY KEY ("id", "egress")
);

CREATE TABLE "Nginx" (
  "id" integer,
  "ingress" integer,
  PRIMARY KEY ("id", "ingress")
);

CREATE TABLE "NginxEgress" (
  "id" integer,
  "egress" integer,
  PRIMARY KEY ("id", "egress")
);

CREATE TABLE "ApplicationDefinitions" (
  "id" integer PRIMARY KEY,
  "name" string,
  "port" integer,
  "type" applicationType
);

CREATE TABLE "ApplicationInstances" (
  "id" integer PRIMARY KEY,
  "server" integer,
  "definition" integer
);

CREATE TABLE "Servers" (
  "id" integer PRIMARY KEY,
  "alias" string,
  "hostname" string
);

CREATE TABLE "Healthchecks" (
  "id" integer PRIMARY KEY,
  "application" integer,
  "url" string,
  "timeout" time,
  "interval" time,
  "expectedstatus" int
);

ALTER TABLE "Proxy" ADD FOREIGN KEY ("id") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "Proxy" ADD FOREIGN KEY ("ingress") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "Proxy" ADD FOREIGN KEY ("egress") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "TopologyNodes" ADD FOREIGN KEY ("id") REFERENCES "F5" ("id");

ALTER TABLE "TopologyNodes" ADD FOREIGN KEY ("id") REFERENCES "F5" ("ingress");

ALTER TABLE "F5Egress" ADD FOREIGN KEY ("id") REFERENCES "F5" ("id");

ALTER TABLE "F5Egress" ADD FOREIGN KEY ("egress") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "Nginx" ADD FOREIGN KEY ("id") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "Nginx" ADD FOREIGN KEY ("ingress") REFERENCES "TopologyNodes" ("id");

ALTER TABLE "NginxEgress" ADD FOREIGN KEY ("id") REFERENCES "Nginx" ("id");

ALTER TABLE "TopologyNodes" ADD FOREIGN KEY ("id") REFERENCES "NginxEgress" ("egress");

ALTER TABLE "ApplicationInstances" ADD FOREIGN KEY ("definition") REFERENCES "ApplicationDefinitions" ("id");

ALTER TABLE "TopologyNodes" ADD FOREIGN KEY ("id") REFERENCES "ApplicationInstances" ("id");

ALTER TABLE "ApplicationInstances" ADD FOREIGN KEY ("server") REFERENCES "Servers" ("id");

ALTER TABLE "ApplicationDefinitions" ADD FOREIGN KEY ("id") REFERENCES "Healthchecks" ("application");
