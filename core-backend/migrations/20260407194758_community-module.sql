-- Modify "community_preferences" table
ALTER TABLE "community_preferences" ADD COLUMN "mute_notifications" boolean NOT NULL DEFAULT false, ADD COLUMN "mute_duration" timestamptz NULL, ADD COLUMN "block_notifications" boolean NOT NULL DEFAULT false;
-- Create "community_categories" table
CREATE TABLE "community_categories" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(200) NOT NULL,
  "slug" character varying(200) NOT NULL,
  "description" text NULL,
  "parent_category_id" uuid NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_community_categories_child_categories" FOREIGN KEY ("parent_category_id") REFERENCES "community_categories" ("id") ON UPDATE CASCADE ON DELETE SET NULL
);
-- Create index "idx_community_categories_deleted_at" to table: "community_categories"
CREATE INDEX "idx_community_categories_deleted_at" ON "community_categories" ("deleted_at");
-- Create index "idx_community_categories_parent" to table: "community_categories"
CREATE INDEX "idx_community_categories_parent" ON "community_categories" ("parent_category_id");
-- Create index "idx_community_categories_slug_per_parent" to table: "community_categories"
CREATE UNIQUE INDEX "idx_community_categories_slug_per_parent" ON "community_categories" ("parent_category_id", "slug");
-- Create "content_reports" table
CREATE TABLE "content_reports" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "reporter_account_id" uuid NOT NULL,
  "target_type" character varying(20) NOT NULL,
  "target_id" uuid NOT NULL,
  "reason" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "admin_note" text NULL,
  "resolved_by_account_id" uuid NULL,
  "resolved_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_content_reports_deleted_at" to table: "content_reports"
CREATE INDEX "idx_content_reports_deleted_at" ON "content_reports" ("deleted_at");
-- Create index "idx_content_reports_reporter" to table: "content_reports"
CREATE INDEX "idx_content_reports_reporter" ON "content_reports" ("reporter_account_id");
-- Create index "idx_content_reports_resolved_by" to table: "content_reports"
CREATE INDEX "idx_content_reports_resolved_by" ON "content_reports" ("resolved_by_account_id");
-- Create index "idx_content_reports_status" to table: "content_reports"
CREATE INDEX "idx_content_reports_status" ON "content_reports" ("status");
-- Create index "idx_content_reports_target" to table: "content_reports"
CREATE INDEX "idx_content_reports_target" ON "content_reports" ("target_type", "target_id");
-- Create "discussion_threads" table
CREATE TABLE "discussion_threads" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "category_id" uuid NOT NULL,
  "author_account_id" uuid NOT NULL,
  "title" character varying(200) NOT NULL,
  "slug" character varying(200) NOT NULL,
  "description" text NULL,
  "is_pinned" boolean NOT NULL DEFAULT false,
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "view_count" bigint NOT NULL DEFAULT 0,
  "share_count" bigint NOT NULL DEFAULT 0,
  "reply_count" bigint NOT NULL DEFAULT 0,
  "last_activity_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_community_categories_threads" FOREIGN KEY ("category_id") REFERENCES "community_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_discussion_threads_author" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_author" ON "discussion_threads" ("author_account_id");
-- Create index "idx_discussion_threads_category" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_category" ON "discussion_threads" ("category_id");
-- Create index "idx_discussion_threads_deleted_at" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_deleted_at" ON "discussion_threads" ("deleted_at");
-- Create index "idx_discussion_threads_last_activity" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_last_activity" ON "discussion_threads" ("last_activity_at");
-- Create index "idx_discussion_threads_slug_per_category" to table: "discussion_threads"
CREATE UNIQUE INDEX "idx_discussion_threads_slug_per_category" ON "discussion_threads" ("category_id", "slug");
-- Create index "idx_discussion_threads_status" to table: "discussion_threads"
CREATE INDEX "idx_discussion_threads_status" ON "discussion_threads" ("status");
-- Create "discussion_posts" table
CREATE TABLE "discussion_posts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "thread_id" uuid NOT NULL,
  "parent_post_id" uuid NULL,
  "author_account_id" uuid NOT NULL,
  "content" text NOT NULL,
  "is_solution" boolean NOT NULL DEFAULT false,
  "is_pinned" boolean NOT NULL DEFAULT false,
  "upvote_count" bigint NOT NULL DEFAULT 0,
  "attachment_url" text NULL,
  "attachment_type" character varying(50) NULL,
  "edit_count" bigint NOT NULL DEFAULT 0,
  "edited_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discussion_posts_replies" FOREIGN KEY ("parent_post_id") REFERENCES "discussion_posts" ("id") ON UPDATE CASCADE ON DELETE SET NULL,
  CONSTRAINT "fk_discussion_threads_posts" FOREIGN KEY ("thread_id") REFERENCES "discussion_threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_discussion_posts_author" to table: "discussion_posts"
CREATE INDEX "idx_discussion_posts_author" ON "discussion_posts" ("author_account_id");
-- Create index "idx_discussion_posts_deleted_at" to table: "discussion_posts"
CREATE INDEX "idx_discussion_posts_deleted_at" ON "discussion_posts" ("deleted_at");
-- Create index "idx_discussion_posts_parent" to table: "discussion_posts"
CREATE INDEX "idx_discussion_posts_parent" ON "discussion_posts" ("parent_post_id");
-- Create index "idx_discussion_posts_solution" to table: "discussion_posts"
CREATE INDEX "idx_discussion_posts_solution" ON "discussion_posts" ("is_solution");
-- Create index "idx_discussion_posts_thread" to table: "discussion_posts"
CREATE INDEX "idx_discussion_posts_thread" ON "discussion_posts" ("thread_id");
-- Create "thread_blocked_users" table
CREATE TABLE "thread_blocked_users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "thread_id" uuid NOT NULL,
  "blocked_account_id" uuid NOT NULL,
  "blocked_by_account_id" uuid NOT NULL,
  "reason" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discussion_threads_blocks" FOREIGN KEY ("thread_id") REFERENCES "discussion_threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_thread_blocked_users_blocked" to table: "thread_blocked_users"
CREATE INDEX "idx_thread_blocked_users_blocked" ON "thread_blocked_users" ("blocked_account_id");
-- Create index "idx_thread_blocked_users_blocked_by" to table: "thread_blocked_users"
CREATE INDEX "idx_thread_blocked_users_blocked_by" ON "thread_blocked_users" ("blocked_by_account_id");
-- Create index "idx_thread_blocked_users_deleted_at" to table: "thread_blocked_users"
CREATE INDEX "idx_thread_blocked_users_deleted_at" ON "thread_blocked_users" ("deleted_at");
-- Create index "idx_thread_blocked_users_thread" to table: "thread_blocked_users"
CREATE INDEX "idx_thread_blocked_users_thread" ON "thread_blocked_users" ("thread_id");
-- Create index "idx_thread_blocked_users_thread_account" to table: "thread_blocked_users"
CREATE UNIQUE INDEX "idx_thread_blocked_users_thread_account" ON "thread_blocked_users" ("thread_id", "blocked_account_id");
-- Create "user_category_settings" table
CREATE TABLE "user_category_settings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "category_id" uuid NOT NULL,
  "is_following" boolean NOT NULL DEFAULT true,
  "is_muted" boolean NOT NULL DEFAULT false,
  "last_read_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_category_settings_category" FOREIGN KEY ("category_id") REFERENCES "community_categories" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_user_category_settings_account" to table: "user_category_settings"
CREATE INDEX "idx_user_category_settings_account" ON "user_category_settings" ("account_id");
-- Create index "idx_user_category_settings_account_category" to table: "user_category_settings"
CREATE UNIQUE INDEX "idx_user_category_settings_account_category" ON "user_category_settings" ("account_id", "category_id");
-- Create index "idx_user_category_settings_category" to table: "user_category_settings"
CREATE INDEX "idx_user_category_settings_category" ON "user_category_settings" ("category_id");
-- Create index "idx_user_category_settings_deleted_at" to table: "user_category_settings"
CREATE INDEX "idx_user_category_settings_deleted_at" ON "user_category_settings" ("deleted_at");
-- Create "user_thread_settings" table
CREATE TABLE "user_thread_settings" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "thread_id" uuid NOT NULL,
  "is_following" boolean NOT NULL DEFAULT true,
  "is_muted" boolean NOT NULL DEFAULT false,
  "last_read_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discussion_threads_followers" FOREIGN KEY ("thread_id") REFERENCES "discussion_threads" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_user_thread_settings_account" to table: "user_thread_settings"
CREATE INDEX "idx_user_thread_settings_account" ON "user_thread_settings" ("account_id");
-- Create index "idx_user_thread_settings_account_thread" to table: "user_thread_settings"
CREATE UNIQUE INDEX "idx_user_thread_settings_account_thread" ON "user_thread_settings" ("account_id", "thread_id");
-- Create index "idx_user_thread_settings_deleted_at" to table: "user_thread_settings"
CREATE INDEX "idx_user_thread_settings_deleted_at" ON "user_thread_settings" ("deleted_at");
-- Create index "idx_user_thread_settings_thread" to table: "user_thread_settings"
CREATE INDEX "idx_user_thread_settings_thread" ON "user_thread_settings" ("thread_id");
