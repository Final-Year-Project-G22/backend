-- Modify "guide_steps" table
ALTER TABLE "guide_steps" ADD COLUMN "compliance_type" character varying(64) NULL;
-- Create "compliance_entries" table
CREATE TABLE "compliance_entries" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "business_profile_id" uuid NOT NULL,
  "account_id" uuid NOT NULL,
  "compliance_type" character varying(64) NOT NULL,
  "reference_number" character varying(255) NULL,
  "issued_date" date NULL,
  "expiry_date" timestamptz NOT NULL,
  "reminder_days_before" bigint NOT NULL DEFAULT 30,
  "source" character varying(20) NOT NULL DEFAULT 'manual',
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "last_notified_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_compliance_account" to table: "compliance_entries"
CREATE INDEX "idx_compliance_account" ON "compliance_entries" ("account_id");
-- Create index "idx_compliance_entries_deleted_at" to table: "compliance_entries"
CREATE INDEX "idx_compliance_entries_deleted_at" ON "compliance_entries" ("deleted_at");
-- Create index "idx_compliance_expiry" to table: "compliance_entries"
CREATE INDEX "idx_compliance_expiry" ON "compliance_entries" ("expiry_date");
-- Create index "idx_compliance_profile" to table: "compliance_entries"
CREATE INDEX "idx_compliance_profile" ON "compliance_entries" ("business_profile_id");
-- Create index "idx_compliance_source" to table: "compliance_entries"
CREATE INDEX "idx_compliance_source" ON "compliance_entries" ("source");
-- Create index "idx_compliance_status" to table: "compliance_entries"
CREATE INDEX "idx_compliance_status" ON "compliance_entries" ("status");
-- Create "compliance_type_localizations" table
CREATE TABLE "compliance_type_localizations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "compliance_type" character varying(64) NOT NULL,
  "locale" character varying(10) NOT NULL,
  "label" character varying(255) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_compliance_type_localizations_deleted_at" to table: "compliance_type_localizations"
CREATE INDEX "idx_compliance_type_localizations_deleted_at" ON "compliance_type_localizations" ("deleted_at");
-- Create index "idx_ct_lang" to table: "compliance_type_localizations"
CREATE UNIQUE INDEX "idx_ct_lang" ON "compliance_type_localizations" ("compliance_type", "locale");
-- Create "scheduled_alert_templates" table
CREATE TABLE "scheduled_alert_templates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "slug" character varying(64) NOT NULL,
  "name" character varying(255) NOT NULL,
  "default_title" character varying(255) NOT NULL,
  "default_body" text NOT NULL,
  "default_channel" character varying(20) NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_scheduled_alert_templates_deleted_at" to table: "scheduled_alert_templates"
CREATE INDEX "idx_scheduled_alert_templates_deleted_at" ON "scheduled_alert_templates" ("deleted_at");
-- Create index "idx_scheduled_alert_templates_slug" to table: "scheduled_alert_templates"
CREATE UNIQUE INDEX "idx_scheduled_alert_templates_slug" ON "scheduled_alert_templates" ("slug");
-- Create "user_scheduled_notifications" table
CREATE TABLE "user_scheduled_notifications" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "template_slug" character varying(64) NULL,
  "title" character varying(255) NOT NULL,
  "body" text NOT NULL,
  "channels" character varying(20)[] NOT NULL,
  "scheduled_for" timestamptz NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "rescheduled_from" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "cancelled_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_scheduled_account" to table: "user_scheduled_notifications"
CREATE INDEX "idx_user_scheduled_account" ON "user_scheduled_notifications" ("account_id");
-- Create index "idx_user_scheduled_notifications_deleted_at" to table: "user_scheduled_notifications"
CREATE INDEX "idx_user_scheduled_notifications_deleted_at" ON "user_scheduled_notifications" ("deleted_at");
-- Create index "idx_user_scheduled_status" to table: "user_scheduled_notifications"
CREATE INDEX "idx_user_scheduled_status" ON "user_scheduled_notifications" ("status");
-- Create index "idx_user_scheduled_time" to table: "user_scheduled_notifications"
CREATE INDEX "idx_user_scheduled_time" ON "user_scheduled_notifications" ("scheduled_for");
