-- Modify "discussion_posts" table
ALTER TABLE "discussion_posts" DROP COLUMN "attachment_url", DROP COLUMN "attachment_type";
-- Create "attachments" table
CREATE TABLE "attachments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "storage_key" text NOT NULL,
  "file_url" text NOT NULL,
  "file_type" character varying(50) NOT NULL,
  "file_name" character varying(255) NOT NULL,
  "file_size" bigint NULL,
  "post_id" uuid NULL,
  "uploaded_by" uuid NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discussion_posts_attachments" FOREIGN KEY ("post_id") REFERENCES "discussion_posts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_attachments_deleted_at" to table: "attachments"
CREATE INDEX "idx_attachments_deleted_at" ON "attachments" ("deleted_at");
-- Create index "idx_attachments_post" to table: "attachments"
CREATE INDEX "idx_attachments_post" ON "attachments" ("post_id");
-- Create index "idx_attachments_uploaded_by" to table: "attachments"
CREATE INDEX "idx_attachments_uploaded_by" ON "attachments" ("uploaded_by");
