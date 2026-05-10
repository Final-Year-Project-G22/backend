-- Modify "notification_campaigns" table
ALTER TABLE "notification_campaigns" DROP COLUMN "template_id", DROP COLUMN "custom_subject", DROP COLUMN "custom_content", ADD COLUMN "campaign_template_id" uuid NOT NULL;
-- Create "campaign_templates" table
CREATE TABLE "campaign_templates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "default_content" jsonb NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_campaign_templates_deleted_at" to table: "campaign_templates"
CREATE INDEX "idx_campaign_templates_deleted_at" ON "campaign_templates" ("deleted_at");
-- Create index "idx_campaign_templates_name" to table: "campaign_templates"
CREATE UNIQUE INDEX "idx_campaign_templates_name" ON "campaign_templates" ("name");
-- Create "campaign_template_translations" table
CREATE TABLE "campaign_template_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "campaign_template_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "content" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_campaign_templates_translations" FOREIGN KEY ("campaign_template_id") REFERENCES "campaign_templates" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_ctrans_template_lang" to table: "campaign_template_translations"
CREATE UNIQUE INDEX "idx_ctrans_template_lang" ON "campaign_template_translations" ("campaign_template_id", "language");
