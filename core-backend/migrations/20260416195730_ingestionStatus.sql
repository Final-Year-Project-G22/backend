-- Create "ingestion_status_events" table
CREATE TABLE "ingestion_status_events" (
  "id" uuid NOT NULL,
  "event_id" character varying(100) NOT NULL,
  "document_id" uuid NOT NULL,
  "account_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "event_type" character varying(100) NOT NULL,
  "schema_version" character varying(32) NOT NULL,
  "occurred_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL,
  "from_stage" character varying(32) NULL,
  "to_stage" character varying(32) NOT NULL,
  "is_terminal" boolean NOT NULL DEFAULT false,
  "retry_count" bigint NOT NULL DEFAULT 0,
  "error_message" text NULL,
  "chunks_processed_count" bigint NOT NULL,
  "chunks_failed_count" bigint NOT NULL,
  "event_sequence" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "idx_ingestion_status_events_event_id" to table: "ingestion_status_events"
CREATE INDEX "idx_ingestion_status_events_event_id" ON "ingestion_status_events" ("event_id");
-- Create index "idx_ingestion_status_events_event_sequence" to table: "ingestion_status_events"
CREATE INDEX "idx_ingestion_status_events_event_sequence" ON "ingestion_status_events" ("event_sequence");
-- Create index "idx_status_account" to table: "ingestion_status_events"
CREATE INDEX "idx_status_account" ON "ingestion_status_events" ("account_id");
-- Create index "idx_status_document_occurred" to table: "ingestion_status_events"
CREATE INDEX "idx_status_document_occurred" ON "ingestion_status_events" ("document_id", "occurred_at");
-- Create index "idx_status_user" to table: "ingestion_status_events"
CREATE INDEX "idx_status_user" ON "ingestion_status_events" ("user_id");
-- Create "ingestion_status_projections" table
CREATE TABLE "ingestion_status_projections" (
  "id" uuid NOT NULL,
  "document_id" uuid NOT NULL,
  "account_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "event_id" character varying(100) NOT NULL,
  "current_stage" character varying(32) NOT NULL,
  "is_terminal" boolean NOT NULL DEFAULT false,
  "started_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "last_error" text NULL,
  "chunks_processed_count" bigint NOT NULL DEFAULT 0,
  "chunks_failed_count" bigint NOT NULL DEFAULT 0,
  "last_event_sequence" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "idx_projection_account" to table: "ingestion_status_projections"
CREATE INDEX "idx_projection_account" ON "ingestion_status_projections" ("account_id");
-- Create index "idx_projection_document" to table: "ingestion_status_projections"
CREATE UNIQUE INDEX "idx_projection_document" ON "ingestion_status_projections" ("document_id");
-- Create index "idx_projection_user" to table: "ingestion_status_projections"
CREATE INDEX "idx_projection_user" ON "ingestion_status_projections" ("user_id");
