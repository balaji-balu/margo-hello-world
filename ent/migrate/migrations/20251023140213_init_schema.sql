-- Create "application_desc" table
CREATE TABLE "application_desc" (
  "app_id" text NOT NULL,
  "name" text NULL,
  "vendor" text NULL,
  "version" text NULL,
  "category" text NULL,
  "description" text NULL,
  "icon" text NULL,
  "artifacturl" text NULL,
  "site" text NULL,
  "tag_line" text NULL,
  "tags" jsonb NULL,
  "published" text NULL,
  PRIMARY KEY ("app_id")
);
-- Create "deployment_profile" table
CREATE TABLE "deployment_profile" (
  "id" text NOT NULL,
  "type" text NULL,
  "description" text NULL,
  "cpu_cores" double precision NULL,
  "memory" text NULL,
  "storage" text NULL,
  "cpu_architectures" jsonb NULL,
  "peripherals" jsonb NULL,
  "interfaces" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create "component" table
CREATE TABLE "component" (
  "id" serial NOT NULL,
  "deployment_profile_id" text NULL,
  "name" text NULL,
  "properties" jsonb NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "component_deployment_profile_id_fkey" FOREIGN KEY ("deployment_profile_id") REFERENCES "deployment_profile" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
