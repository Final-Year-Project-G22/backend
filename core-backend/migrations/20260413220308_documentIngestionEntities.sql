-- Create "ingestion_documents" table
CREATE TABLE "ingestion_documents" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "storage_key" text NOT NULL,
  "content_type" character varying(255) NOT NULL,
  "size_bytes" bigint NOT NULL DEFAULT 0,
  "checksum_sha256" character varying(128) NOT NULL,
  "idempotency_key" character varying(255) NOT NULL,
  "batch_id" uuid NULL,
  "source_filename" text NULL,
  "declared_language" character varying(16) NULL,
  "schema_version" character varying(32) NOT NULL DEFAULT '1.0.0',
  "status" character varying(32) NOT NULL DEFAULT 'queued',
  "last_error" text NULL,
  "event_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_ingestion_docs_idempotency_per_account" to table: "ingestion_documents"
CREATE UNIQUE INDEX "idx_ingestion_docs_idempotency_per_account" ON "ingestion_documents" ("account_id", "idempotency_key");
-- Create index "idx_ingestion_documents_account" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_account" ON "ingestion_documents" ("account_id");
-- Create index "idx_ingestion_documents_batch" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_batch" ON "ingestion_documents" ("batch_id");
-- Create index "idx_ingestion_documents_deleted_at" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_deleted_at" ON "ingestion_documents" ("deleted_at");
-- Create index "idx_ingestion_documents_event_id" to table: "ingestion_documents"
CREATE UNIQUE INDEX "idx_ingestion_documents_event_id" ON "ingestion_documents" ("event_id");
-- Create index "idx_ingestion_documents_status" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_status" ON "ingestion_documents" ("status");
-- Create index "idx_ingestion_documents_storage_key" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_storage_key" ON "ingestion_documents" ("storage_key");
-- Create index "idx_ingestion_documents_user" to table: "ingestion_documents"
CREATE INDEX "idx_ingestion_documents_user" ON "ingestion_documents" ("user_id");
-- Create "ingestion_outbox" table
CREATE TABLE "ingestion_outbox" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "event_id" uuid NOT NULL,
  "event_type" character varying(255) NOT NULL,
  "schema_version" character varying(32) NOT NULL DEFAULT '1.0.0',
  "producer" character varying(64) NOT NULL,
  "key_id" character varying(128) NOT NULL,
  "idempotency_key" character varying(255) NOT NULL,
  "aggregate_id" uuid NOT NULL,
  "account_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "batch_id" uuid NULL,
  "replay_count" integer NOT NULL DEFAULT 0,
  "payload" jsonb NOT NULL,
  "signature" bytea NULL,
  "status" character varying(32) NOT NULL DEFAULT 'pending',
  "attempt_count" bigint NOT NULL DEFAULT 0,
  "next_attempt_at" timestamptz NULL,
  "published_at" timestamptz NULL,
  "last_error" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_ingestion_outbox_account" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_account" ON "ingestion_outbox" ("account_id");
-- Create index "idx_ingestion_outbox_aggregate" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_aggregate" ON "ingestion_outbox" ("aggregate_id");
-- Create index "idx_ingestion_outbox_batch" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_batch" ON "ingestion_outbox" ("batch_id");
-- Create index "idx_ingestion_outbox_dedupe" to table: "ingestion_outbox"
CREATE UNIQUE INDEX "idx_ingestion_outbox_dedupe" ON "ingestion_outbox" ("event_type", "idempotency_key");
-- Create index "idx_ingestion_outbox_deleted_at" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_deleted_at" ON "ingestion_outbox" ("deleted_at");
-- Create index "idx_ingestion_outbox_event_id" to table: "ingestion_outbox"
CREATE UNIQUE INDEX "idx_ingestion_outbox_event_id" ON "ingestion_outbox" ("event_id");
-- Create index "idx_ingestion_outbox_event_type" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_event_type" ON "ingestion_outbox" ("event_type");
-- Create index "idx_ingestion_outbox_status_next_attempt" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_status_next_attempt" ON "ingestion_outbox" ("status", "next_attempt_at");
-- Create index "idx_ingestion_outbox_user" to table: "ingestion_outbox"
CREATE INDEX "idx_ingestion_outbox_user" ON "ingestion_outbox" ("user_id");
