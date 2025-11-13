CREATE TABLE "application_desc" (
  "app_id" text PRIMARY KEY,
  "name" text,
  "vendor" text,
  "version" text,
  "category" text,
  "description" text,
  "icon" text,
  "artifacturl" text,
  "site" text,
  "tag_line" text,
  "tags" JSONB,
  "published" text
);

CREATE TABLE "deployment_profile" (
  "id" TEXT PRIMARY KEY,
  "type" TEXT,
  "description" TEXT,
  "cpu_cores" FLOAT,
  "memory" TEXT,
  "storage" TEXT,
  "cpu_architectures" JSONB,
  "peripherals" JSONB,
  "interfaces" JSONB,
  "app_id" text
);

CREATE TABLE "component" (
  "id" SERIAL PRIMARY KEY,
  "deployment_profile_id" TEXT,
  "name" TEXT,
  "properties" JSONB
);

CREATE TABLE "site" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "site_id" varchar UNIQUE NOT NULL,
  "name" varchar,
  "description" text,
  "location" varchar,
  "orchestrator_id" uuid,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "host" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "host_id" varchar UNIQUE NOT NULL,
  "site_id" uuid NOT NULL,
  "hostname" varchar,
  "ip_address" text,
  "edge_url" varchar,
  "status" varchar,
  "last_heartbeat" timestamptz,
  "metadata" jsonb,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "orchestrator" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" varchar,
  "type" varchar,
  "region" varchar,
  "api_endpoint" varchar,
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now())
);

CREATE TABLE "deployment_status" {
  "deployment_id" varchar,
  "status" varchar
  "components"
  "created_at" timestamptz DEFAULT (now()),
  "updated_at" timestamptz DEFAULT (now()) 
}

ALTER TABLE "deployment_profile" ADD FOREIGN KEY ("app_id") REFERENCES "application_desc" ("app_id");

ALTER TABLE "component" ADD FOREIGN KEY ("deployment_profile_id") REFERENCES "deployment_profile" ("id");

ALTER TABLE "site" ADD FOREIGN KEY ("orchestrator_id") REFERENCES "orchestrator" ("id");

ALTER TABLE "host" ADD FOREIGN KEY ("site_id") REFERENCES "site" ("id");
