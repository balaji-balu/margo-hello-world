-- Create "orchestrator" table
CREATE TABLE "orchestrator" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying NULL,
  "type" character varying NULL,
  "region" character varying NULL,
  "api_endpoint" character varying NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create "site" table
CREATE TABLE "site" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "site_id" character varying NOT NULL,
  "name" character varying NULL,
  "description" text NULL,
  "location" character varying NULL,
  "orchestrator_id" uuid NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "site_site_id_key" UNIQUE ("site_id"),
  CONSTRAINT "site_orchestrator_id_fkey" FOREIGN KEY ("orchestrator_id") REFERENCES "orchestrator" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "host" table
CREATE TABLE "host" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "host_id" character varying NOT NULL,
  "site_id" uuid NOT NULL,
  "hostname" character varying NULL,
  "ip_address" inet NULL,
  "edge_url" character varying NULL,
  "status" character varying NULL,
  "last_heartbeat" timestamptz NULL,
  "metadata" jsonb NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "host_host_id_key" UNIQUE ("host_id"),
  CONSTRAINT "host_site_id_fkey" FOREIGN KEY ("site_id") REFERENCES "site" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
