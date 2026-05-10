-- Modify "business_profiles" table
ALTER TABLE "business_profiles" DROP COLUMN "business_type", DROP COLUMN "business_sector", DROP COLUMN "location", ADD COLUMN "physical_address" character varying(255) NULL, ADD COLUMN "region" character varying(50) NULL, ADD COLUMN "stage" character varying(50) NULL, ADD COLUMN "sector_id" uuid NULL;
-- Create index "idx_business_profiles_region" to table: "business_profiles"
CREATE INDEX "idx_business_profiles_region" ON "business_profiles" ("region");
-- Create index "idx_business_profiles_sector_id" to table: "business_profiles"
CREATE INDEX "idx_business_profiles_sector_id" ON "business_profiles" ("sector_id");
-- Create index "idx_business_profiles_stage" to table: "business_profiles"
CREATE INDEX "idx_business_profiles_stage" ON "business_profiles" ("stage");
-- Modify "discussion_threads" table
ALTER TABLE "discussion_threads" DROP COLUMN "category_id", ADD COLUMN "sector_ids" uuid[] NULL, ADD COLUMN "tag_ids" uuid[] NULL;
-- Create index "idx_discussion_threads_sector_ids" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_sector_ids" ON "discussion_threads" ("sector_ids");
-- Create index "idx_discussion_threads_slug" to table: "discussion_threads"
CREATE UNIQUE INDEX "idx_discussion_threads_slug" ON "discussion_threads" ("slug");
-- Create index "idx_discussion_threads_tag_ids" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_tag_ids" ON "discussion_threads" ("tag_ids");
-- Modify "guide_category_conditions" table
ALTER TABLE "guide_category_conditions" DROP CONSTRAINT "fk_guide_categories_conditions";
-- Modify "guide_category_translations" table
ALTER TABLE "guide_category_translations" DROP CONSTRAINT "fk_guide_categories_translations";
-- Modify "guides" table
ALTER TABLE "guides" DROP COLUMN "category_id", ADD COLUMN "sector_ids" uuid[] NULL, ADD COLUMN "tag_ids" uuid[] NULL;
-- Create index "idx_guides_sector_ids" to table: "guides"
CREATE INDEX "idx_guides_sector_ids" ON "guides" ("sector_ids");
-- Create index "idx_guides_slug" to table: "guides"
CREATE UNIQUE INDEX "idx_guides_slug" ON "guides" ("slug");
-- Create index "idx_guides_tag_ids" to table: "guides"
CREATE INDEX "idx_guides_tag_ids" ON "guides" ("tag_ids");
-- Modify "ingestion_documents" table
ALTER TABLE "ingestion_documents" ADD COLUMN "sector_ids" uuid[] NULL, ADD COLUMN "tag_ids" uuid[] NULL, ADD COLUMN "region" character varying(50) NULL, ADD COLUMN "stage" character varying(50) NULL;
-- Create index "idx_ingestion_documents_sector_ids" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_sector_ids" ON "ingestion_documents" ("sector_ids");
-- Create index "idx_ingestion_documents_tag_ids" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_tag_ids" ON "ingestion_documents" ("tag_ids");
-- Modify "library_template_groups" table
ALTER TABLE "library_template_groups" DROP CONSTRAINT "fk_library_categories_template_groups";
-- Modify "notification_campaigns" table
ALTER TABLE "notification_campaigns" ADD COLUMN "sector_ids" uuid[] NULL, ADD COLUMN "tag_ids" uuid[] NULL, ADD COLUMN "region" character varying(50) NULL, ADD COLUMN "stage" character varying(50) NULL;
-- Create index "idx_notif_campaigns_sector_ids" to table: "notification_campaigns"
CREATE INDEX "idx_notif_campaigns_sector_ids" ON "notification_campaigns" ("sector_ids");
-- Create index "idx_notif_campaigns_tag_ids" to table: "notification_campaigns"
CREATE INDEX "idx_notif_campaigns_tag_ids" ON "notification_campaigns" ("tag_ids");
-- Modify "notification_templates" table
ALTER TABLE "notification_templates" DROP COLUMN "category", ADD COLUMN "template_group" character varying(100) NULL;
-- Create index "idx_notif_templates_group" to table: "notification_templates"
CREATE INDEX "idx_notif_templates_group" ON "notification_templates" ("template_group");
-- Modify "user_guide_progresses" table
ALTER TABLE "user_guide_progresses" ALTER COLUMN "uploaded_documents" SET DEFAULT '[]';
-- Modify "user_notification_inboxes" table
ALTER TABLE "user_notification_inboxes" DROP COLUMN "category";
-- Create "business_profile_tags" table
CREATE TABLE "business_profile_tags" (
  "business_profile_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tag_id" uuid NOT NULL DEFAULT gen_random_uuid(),
  PRIMARY KEY ("business_profile_id", "tag_id")
);
-- Create "sector_translations" table
CREATE TABLE "sector_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "sector_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create index "idx_sector_trans" to table: "sector_translations"
CREATE UNIQUE INDEX "idx_sector_trans" ON "sector_translations" ("sector_id", "language");
-- Create "sectors" table
CREATE TABLE "sectors" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "slug" character varying(100) NOT NULL,
  "parent_id" uuid NULL,
  "ancestor_ids" uuid[] NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_sectors_parent" FOREIGN KEY ("parent_id") REFERENCES "sectors" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sectors_ancestor_ids" to table: "sectors"
CREATE INDEX "idx_sectors_ancestor_ids" ON "sectors" ("ancestor_ids");
-- Create index "idx_sectors_deleted_at" to table: "sectors"
CREATE INDEX "idx_sectors_deleted_at" ON "sectors" ("deleted_at");
-- Create index "idx_sectors_parent_id" to table: "sectors"
CREATE INDEX "idx_sectors_parent_id" ON "sectors" ("parent_id");
-- Create index "idx_sectors_slug" to table: "sectors"
CREATE UNIQUE INDEX "idx_sectors_slug" ON "sectors" ("slug");
-- Create "tag_translations" table
CREATE TABLE "tag_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "tag_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id")
);
-- Create index "idx_tag_trans" to table: "tag_translations"
CREATE UNIQUE INDEX "idx_tag_trans" ON "tag_translations" ("tag_id", "language");
-- Create "tags" table
CREATE TABLE "tags" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "slug" character varying(100) NOT NULL,
  "group" character varying(50) NOT NULL,
  "is_multi_select" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_tags_deleted_at" to table: "tags"
CREATE INDEX "idx_tags_deleted_at" ON "tags" ("deleted_at");
-- Create index "idx_tags_group" to table: "tags"
CREATE INDEX "idx_tags_group" ON "tags" ("group");
-- Create index "idx_tags_slug" to table: "tags"
CREATE UNIQUE INDEX "idx_tags_slug" ON "tags" ("slug");
-- Modify "business_profiles" table
ALTER TABLE "business_profiles" ADD CONSTRAINT "fk_business_profiles_sector" FOREIGN KEY ("sector_id") REFERENCES "sectors" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
-- Modify "library_template_groups" table
ALTER TABLE "library_template_groups" ADD CONSTRAINT "fk_library_template_groups_category" FOREIGN KEY ("category_id") REFERENCES "library_categories" ("id") ON UPDATE CASCADE ON DELETE RESTRICT;
-- Modify "business_profile_tags" table
ALTER TABLE "business_profile_tags" ADD CONSTRAINT "fk_business_profile_tags_business_profile" FOREIGN KEY ("business_profile_id") REFERENCES "business_profiles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD CONSTRAINT "fk_business_profile_tags_tag" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION;
-- Modify "sector_translations" table
ALTER TABLE "sector_translations" ADD CONSTRAINT "fk_sectors_translations" FOREIGN KEY ("sector_id") REFERENCES "sectors" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Modify "tag_translations" table
ALTER TABLE "tag_translations" ADD CONSTRAINT "fk_tags_translations" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE CASCADE ON DELETE CASCADE;
-- Drop "guide_categories" table
DROP TABLE "guide_categories";
-- Drop "guide_category_conditions" table
DROP TABLE "guide_category_conditions";
-- Drop "guide_category_translations" table
DROP TABLE "guide_category_translations";
