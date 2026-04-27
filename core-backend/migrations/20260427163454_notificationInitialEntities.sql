-- Create "email_delivery_logs" table
CREATE TABLE "email_delivery_logs" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "notification_history_id" uuid NOT NULL,
  "provider" character varying(50) NOT NULL,
  "provider_message_id" character varying(255) NULL,
  "recipient_email" character varying(255) NOT NULL,
  "subject" character varying(500) NOT NULL,
  "sent_at" timestamptz NOT NULL,
  "delivered_at" timestamptz NULL,
  "opened_at" timestamptz NULL,
  "clicked_at" timestamptz NULL,
  "bounce_reason" text NULL,
  "complaint" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "idx_email_delivery_logs_deleted_at" to table: "email_delivery_logs"
CREATE INDEX "idx_email_delivery_logs_deleted_at" ON "email_delivery_logs" ("deleted_at");
-- Create index "idx_email_delivery_provider_msg" to table: "email_delivery_logs"
CREATE INDEX "idx_email_delivery_provider_msg" ON "email_delivery_logs" ("provider_message_id");
-- Create "muted_accounts" table
CREATE TABLE "muted_accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "muted_account_id" uuid NOT NULL,
  "mute_until" timestamptz NULL,
  "reason" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_muted_accounts_account" to table: "muted_accounts"
CREATE INDEX "idx_muted_accounts_account" ON "muted_accounts" ("account_id");
-- Create index "idx_muted_accounts_account_pair" to table: "muted_accounts"
CREATE UNIQUE INDEX "idx_muted_accounts_account_pair" ON "muted_accounts" ("account_id", "muted_account_id");
-- Create index "idx_muted_accounts_deleted_at" to table: "muted_accounts"
CREATE INDEX "idx_muted_accounts_deleted_at" ON "muted_accounts" ("deleted_at");
-- Create index "idx_muted_accounts_muted" to table: "muted_accounts"
CREATE INDEX "idx_muted_accounts_muted" ON "muted_accounts" ("muted_account_id");
-- Create "notification_campaigns" table
CREATE TABLE "notification_campaigns" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "campaign_type" character varying(20) NOT NULL,
  "target_segment" jsonb NULL,
  "template_id" uuid NOT NULL,
  "custom_subject" character varying(500) NULL,
  "custom_content" jsonb NULL,
  "scheduled_for" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "created_by" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notif_campaigns_creator" to table: "notification_campaigns"
CREATE INDEX "idx_notif_campaigns_creator" ON "notification_campaigns" ("created_by");
-- Create index "idx_notification_campaigns_deleted_at" to table: "notification_campaigns"
CREATE INDEX "idx_notification_campaigns_deleted_at" ON "notification_campaigns" ("deleted_at");
-- Create "notification_queue" table
CREATE TABLE "notification_queue" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "notification_type" character varying(64) NOT NULL,
  "account_id" uuid NOT NULL,
  "priority" smallint NOT NULL DEFAULT 1,
  "template_id" uuid NULL,
  "channel" character varying(20) NOT NULL,
  "payload" jsonb NOT NULL,
  "scheduled_for" timestamptz NOT NULL,
  "max_retries" bigint NOT NULL DEFAULT 3,
  "retry_count" bigint NOT NULL DEFAULT 0,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "error_message" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notif_queue_account" to table: "notification_queue"
CREATE INDEX "idx_notif_queue_account" ON "notification_queue" ("account_id");
-- Create index "idx_notif_queue_scheduled" to table: "notification_queue"
CREATE INDEX "idx_notif_queue_scheduled" ON "notification_queue" ("scheduled_for");
-- Create index "idx_notif_queue_status" to table: "notification_queue"
CREATE INDEX "idx_notif_queue_status" ON "notification_queue" ("status");
-- Create index "idx_notification_queue_deleted_at" to table: "notification_queue"
CREATE INDEX "idx_notification_queue_deleted_at" ON "notification_queue" ("deleted_at");
-- Create "notification_templates" table
CREATE TABLE "notification_templates" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "description" text NULL,
  "notification_type" character varying(64) NOT NULL,
  "category" character varying(32) NOT NULL,
  "priority" smallint NOT NULL DEFAULT 1,
  "is_system_managed" boolean NOT NULL DEFAULT false,
  "default_content" jsonb NOT NULL,
  "variables_schema" jsonb NULL,
  "default_ttl" integer NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notif_templates_category" to table: "notification_templates"
CREATE INDEX "idx_notif_templates_category" ON "notification_templates" ("category");
-- Create index "idx_notif_templates_name" to table: "notification_templates"
CREATE UNIQUE INDEX "idx_notif_templates_name" ON "notification_templates" ("name");
-- Create index "idx_notif_templates_type" to table: "notification_templates"
CREATE UNIQUE INDEX "idx_notif_templates_type" ON "notification_templates" ("notification_type");
-- Create index "idx_notification_templates_deleted_at" to table: "notification_templates"
CREATE INDEX "idx_notification_templates_deleted_at" ON "notification_templates" ("deleted_at");
-- Create "user_devices" table
CREATE TABLE "user_devices" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "device_type" character varying(20) NOT NULL,
  "device_token" character varying(512) NOT NULL,
  "push_token" text NULL,
  "device_name" character varying(200) NULL,
  "device_model" character varying(200) NULL,
  "os_version" character varying(50) NULL,
  "app_version" character varying(50) NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "last_active_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_devices_account" to table: "user_devices"
CREATE INDEX "idx_user_devices_account" ON "user_devices" ("account_id");
-- Create index "idx_user_devices_deleted_at" to table: "user_devices"
CREATE INDEX "idx_user_devices_deleted_at" ON "user_devices" ("deleted_at");
-- Create index "idx_user_devices_token" to table: "user_devices"
CREATE UNIQUE INDEX "idx_user_devices_token" ON "user_devices" ("device_token");
-- Create "user_notification_preferences" table
CREATE TABLE "user_notification_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "notification_type" character varying(64) NOT NULL,
  "channel" character varying(20) NOT NULL,
  "is_enabled" boolean NOT NULL,
  "quiet_hours_start" timestamptz NULL,
  "quiet_hours_end" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_notif_prefs_account" to table: "user_notification_preferences"
CREATE INDEX "idx_user_notif_prefs_account" ON "user_notification_preferences" ("account_id");
-- Create index "idx_user_notif_prefs_account_type_channel" to table: "user_notification_preferences"
CREATE UNIQUE INDEX "idx_user_notif_prefs_account_type_channel" ON "user_notification_preferences" ("account_id", "notification_type", "channel");
-- Create index "idx_user_notification_preferences_deleted_at" to table: "user_notification_preferences"
CREATE INDEX "idx_user_notification_preferences_deleted_at" ON "user_notification_preferences" ("deleted_at");
-- Create "notification_template_translations" table
CREATE TABLE "notification_template_translations" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "template_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL,
  "subject" character varying(500) NOT NULL,
  "content" jsonb NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notification_templates_translations" FOREIGN KEY ("template_id") REFERENCES "notification_templates" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_notif_template_trans" to table: "notification_template_translations"
CREATE UNIQUE INDEX "idx_notif_template_trans" ON "notification_template_translations" ("template_id", "language");
-- Create "notification_histories" table
CREATE TABLE "notification_histories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "notification_type" character varying(64) NOT NULL,
  "channel" character varying(20) NOT NULL,
  "title" character varying(500) NOT NULL,
  "content" text NOT NULL,
  "action_url" character varying(512) NULL,
  "sent_at" timestamptz NOT NULL,
  "delivered_at" timestamptz NULL,
  "read_at" timestamptz NULL,
  "clicked_at" timestamptz NULL,
  "delivery_status" character varying(20) NOT NULL,
  "failure_reason" text NULL,
  "metadata" jsonb NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notif_history_account" to table: "notification_histories"
CREATE INDEX "idx_notif_history_account" ON "notification_histories" ("account_id");
-- Create index "idx_notification_histories_deleted_at" to table: "notification_histories"
CREATE INDEX "idx_notification_histories_deleted_at" ON "notification_histories" ("deleted_at");
-- Create "user_notification_inboxes" table
CREATE TABLE "user_notification_inboxes" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "notification_history_id" uuid NOT NULL,
  "category" character varying(32) NOT NULL,
  "action_url" character varying(512) NULL,
  "is_read" boolean NOT NULL DEFAULT false,
  "is_archived" boolean NOT NULL DEFAULT false,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_notification_inboxes_notification_history" FOREIGN KEY ("notification_history_id") REFERENCES "notification_histories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_notif_inbox_account" to table: "user_notification_inboxes"
CREATE INDEX "idx_notif_inbox_account" ON "user_notification_inboxes" ("account_id");
-- Create index "idx_notif_inbox_category" to table: "user_notification_inboxes"
CREATE INDEX "idx_notif_inbox_category" ON "user_notification_inboxes" ("category");
-- Create index "idx_notif_inbox_expires" to table: "user_notification_inboxes"
CREATE INDEX "idx_notif_inbox_expires" ON "user_notification_inboxes" ("expires_at");
-- Create index "idx_user_notification_inboxes_deleted_at" to table: "user_notification_inboxes"
CREATE INDEX "idx_user_notification_inboxes_deleted_at" ON "user_notification_inboxes" ("deleted_at");
