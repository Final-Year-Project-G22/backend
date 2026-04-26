-- Modify "content_reports" table
ALTER TABLE "content_reports" DROP COLUMN "target_type", DROP COLUMN "target_id", ADD COLUMN "thread_id" uuid NULL, ADD COLUMN "post_id" uuid NULL, ADD COLUMN "reported_account_id" uuid NULL;
-- Create index "idx_content_reports_post" to table: "content_reports"
CREATE INDEX "idx_content_reports_post" ON "content_reports" ("post_id");
-- Create index "idx_content_reports_reported_account" to table: "content_reports"
CREATE INDEX "idx_content_reports_reported_account" ON "content_reports" ("reported_account_id");
-- Create index "idx_content_reports_thread" to table: "content_reports"
CREATE INDEX "idx_content_reports_thread" ON "content_reports" ("thread_id");
-- Modify "discussion_threads" table
ALTER TABLE "discussion_threads" ADD COLUMN "parent_thread_id" uuid NULL, ADD CONSTRAINT "fk_discussion_threads_sub_threads" FOREIGN KEY ("parent_thread_id") REFERENCES "discussion_threads" ("id") ON UPDATE CASCADE ON DELETE SET NULL;
-- Create index "idx_threads_parent" to table: "discussion_threads"
CREATE INDEX "idx_threads_parent" ON "discussion_threads" ("parent_thread_id");
-- Create index "idx_threads_slug_per_parent" to table: "discussion_threads"
CREATE UNIQUE INDEX "idx_threads_slug_per_parent" ON "discussion_threads" ("parent_thread_id");
