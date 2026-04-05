-- Modify "accounts" table
ALTER TABLE "accounts" ADD COLUMN "username" character varying(64) NULL, ADD COLUMN "username_normalized" character varying(64) NULL;
-- Create index "idx_accounts_username_normalized" to table: "accounts"
CREATE UNIQUE INDEX "idx_accounts_username_normalized" ON "accounts" ("username_normalized");
-- Drop index "idx_bookmark_user_step" from table: "user_guide_bookmarks"
DROP INDEX "idx_bookmark_user_step";
-- Modify "user_guide_bookmarks" table
ALTER TABLE "user_guide_bookmarks" ADD COLUMN "account_id" uuid NOT NULL;
-- Create index "idx_bookmark_account_user_step" to table: "user_guide_bookmarks"
CREATE UNIQUE INDEX "idx_bookmark_account_user_step" ON "user_guide_bookmarks" ("account_id", "user_id", "step_id");
-- Create index "idx_bookmarks_account" to table: "user_guide_bookmarks"
CREATE INDEX "idx_bookmarks_account" ON "user_guide_bookmarks" ("account_id");
-- Drop index "idx_journey_user_guide" from table: "user_guide_journeys"
DROP INDEX "idx_journey_user_guide";
-- Modify "user_guide_journeys" table
ALTER TABLE "user_guide_journeys" ADD COLUMN "account_id" uuid NOT NULL;
-- Create index "idx_journey_account" to table: "user_guide_journeys"
CREATE INDEX "idx_journey_account" ON "user_guide_journeys" ("account_id");
-- Create index "idx_journey_account_user_guide" to table: "user_guide_journeys"
CREATE UNIQUE INDEX "idx_journey_account_user_guide" ON "user_guide_journeys" ("account_id", "user_id", "guide_id");
-- Modify "user_guide_progresses" table
ALTER TABLE "user_guide_progresses" ADD COLUMN "account_id" uuid NOT NULL;
-- Create index "idx_user_progress_account" to table: "user_guide_progresses"
CREATE INDEX "idx_user_progress_account" ON "user_guide_progresses" ("account_id");
-- Create index "idx_user_progress_account_user_step" to table: "user_guide_progresses"
CREATE UNIQUE INDEX "idx_user_progress_account_user_step" ON "user_guide_progresses" ("account_id", "user_id", "step_id");
