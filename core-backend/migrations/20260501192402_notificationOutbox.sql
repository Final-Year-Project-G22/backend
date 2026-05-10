-- Create "notification_outbox" table
CREATE TABLE "notification_outbox" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "event_type" character varying(255) NOT NULL,
  "schema_version" character varying(32) NOT NULL DEFAULT '1.0.0',
  "source_module" character varying(64) NOT NULL,
  "account_id" uuid NOT NULL,
  "idempotency_key" character varying(255) NOT NULL,
  "payload" jsonb NOT NULL,
  "status" character varying(32) NOT NULL DEFAULT 'pending',
  "attempt_count" bigint NOT NULL DEFAULT 0,
  "next_attempt_at" timestamptz NULL,
  "published_at" timestamptz NULL,
  "last_error" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_notification_outbox_account" to table: "notification_outbox"
CREATE INDEX "idx_notification_outbox_account" ON "notification_outbox" ("account_id");
-- Create index "idx_notification_outbox_deleted_at" to table: "notification_outbox"
CREATE INDEX "idx_notification_outbox_deleted_at" ON "notification_outbox" ("deleted_at");
-- Create index "idx_notification_outbox_idempotency" to table: "notification_outbox"
CREATE UNIQUE INDEX "idx_notification_outbox_idempotency" ON "notification_outbox" ("idempotency_key");
-- Create index "idx_notification_outbox_source" to table: "notification_outbox"
CREATE INDEX "idx_notification_outbox_source" ON "notification_outbox" ("source_module");
-- Create index "idx_notification_outbox_status_next" to table: "notification_outbox"
CREATE INDEX "idx_notification_outbox_status_next" ON "notification_outbox" ("status", "next_attempt_at");
