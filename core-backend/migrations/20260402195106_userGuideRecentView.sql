-- Create "user_guide_recent_views" table
CREATE TABLE "user_guide_recent_views" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "guide_id" uuid NOT NULL,
  "last_viewed_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "view_count" bigint NOT NULL DEFAULT 1,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_guide_recent_views_guide" FOREIGN KEY ("guide_id") REFERENCES "guides" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_recent_view_account" to table: "user_guide_recent_views"
CREATE INDEX "idx_recent_view_account" ON "user_guide_recent_views" ("account_id");
-- Create index "idx_recent_view_account_user_guide" to table: "user_guide_recent_views"
CREATE UNIQUE INDEX "idx_recent_view_account_user_guide" ON "user_guide_recent_views" ("account_id", "user_id", "guide_id");
-- Create index "idx_recent_view_guide" to table: "user_guide_recent_views"
CREATE INDEX "idx_recent_view_guide" ON "user_guide_recent_views" ("guide_id");
-- Create index "idx_recent_view_last_viewed" to table: "user_guide_recent_views"
CREATE INDEX "idx_recent_view_last_viewed" ON "user_guide_recent_views" ("last_viewed_at");
-- Create index "idx_recent_view_user" to table: "user_guide_recent_views"
CREATE INDEX "idx_recent_view_user" ON "user_guide_recent_views" ("user_id");
-- Create index "idx_user_guide_recent_views_deleted_at" to table: "user_guide_recent_views"
CREATE INDEX "idx_user_guide_recent_views_deleted_at" ON "user_guide_recent_views" ("deleted_at");
