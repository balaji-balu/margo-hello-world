-- Modify "deployment_profile" table
ALTER TABLE "deployment_profile" ADD COLUMN "app_id" text NULL, ADD CONSTRAINT "deployment_profile_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "application_desc" ("app_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
