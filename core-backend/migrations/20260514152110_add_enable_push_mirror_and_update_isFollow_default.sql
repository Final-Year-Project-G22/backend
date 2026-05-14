-- Modify "campaign_templates" table
ALTER TABLE "campaign_templates" ADD COLUMN "enable_push_mirror" boolean NOT NULL DEFAULT false;
-- Modify "user_thread_settings" table
ALTER TABLE "user_thread_settings" ALTER COLUMN "is_following" SET DEFAULT false;
