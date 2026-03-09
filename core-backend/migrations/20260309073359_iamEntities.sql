-- Create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "first_name" character varying(100) NOT NULL,
  "last_name" character varying(100) NOT NULL,
  "image_url" character varying(512) NULL,
  "bio" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "users" ("deleted_at");
-- Create "accounts" table
CREATE TABLE "accounts" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  "email" character varying(255) NOT NULL,
  "email_normalized" character varying(255) NOT NULL,
  "password_hash" character varying(255) NOT NULL,
  "phone_number" character varying(50) NULL,
  "email_verified" boolean NOT NULL DEFAULT false,
  "phone_verified" boolean NOT NULL DEFAULT false,
  "status" character varying(64) NOT NULL DEFAULT 'pending_verification',
  "failed_login_attempts" bigint NOT NULL DEFAULT 0,
  "locked_until" timestamptz NULL,
  "last_login_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_accounts" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_accounts_deleted_at" to table: "accounts"
CREATE INDEX "idx_accounts_deleted_at" ON "accounts" ("deleted_at");
-- Create index "idx_accounts_email_normalized" to table: "accounts"
CREATE UNIQUE INDEX "idx_accounts_email_normalized" ON "accounts" ("email_normalized");
-- Create index "idx_accounts_user_id" to table: "accounts"
CREATE INDEX "idx_accounts_user_id" ON "accounts" ("user_id");
-- Create "account_preferences" table
CREATE TABLE "account_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "language" character varying(10) NOT NULL DEFAULT 'en',
  "timezone" character varying(64) NOT NULL DEFAULT 'UTC',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_account_preference" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_account_preferences_account_id" to table: "account_preferences"
CREATE UNIQUE INDEX "idx_account_preferences_account_id" ON "account_preferences" ("account_id");
-- Create index "idx_account_preferences_deleted_at" to table: "account_preferences"
CREATE INDEX "idx_account_preferences_deleted_at" ON "account_preferences" ("deleted_at");
-- Create "ai_preferences" table
CREATE TABLE "ai_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "default_model" character varying(100) NULL,
  "response_style" character varying(100) NULL,
  "temperature" numeric NULL,
  "allow_data_retention" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_a_ipreference" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_ai_preferences_account_id" to table: "ai_preferences"
CREATE UNIQUE INDEX "idx_ai_preferences_account_id" ON "ai_preferences" ("account_id");
-- Create index "idx_ai_preferences_deleted_at" to table: "ai_preferences"
CREATE INDEX "idx_ai_preferences_deleted_at" ON "ai_preferences" ("deleted_at");
-- Create "business_profiles" table
CREATE TABLE "business_profiles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "company_name" character varying(255) NOT NULL,
  "company_email" character varying(255) NOT NULL,
  "company_phone_number" character varying(50) NOT NULL,
  "business_type" character varying(100) NULL,
  "business_sector" character varying(100) NULL,
  "registration_number" character varying(100) NULL,
  "registration_date" date NULL,
  "tax_identification_number" character varying(100) NULL,
  "trade_license_number" character varying(100) NULL,
  "location" character varying(255) NULL,
  "description" text NULL,
  "logo_url" character varying(512) NULL,
  "banner_url" character varying(512) NULL,
  "social_links" jsonb NOT NULL DEFAULT '{}',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_business_profile" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_business_profiles_account_id" to table: "business_profiles"
CREATE UNIQUE INDEX "idx_business_profiles_account_id" ON "business_profiles" ("account_id");
-- Create index "idx_business_profiles_deleted_at" to table: "business_profiles"
CREATE INDEX "idx_business_profiles_deleted_at" ON "business_profiles" ("deleted_at");
-- Create "community_preferences" table
CREATE TABLE "community_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "allow_mentions" boolean NOT NULL DEFAULT true,
  "allow_replies" boolean NOT NULL DEFAULT true,
  "digest_enabled" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_community_preference" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_community_preferences_account_id" to table: "community_preferences"
CREATE UNIQUE INDEX "idx_community_preferences_account_id" ON "community_preferences" ("account_id");
-- Create index "idx_community_preferences_deleted_at" to table: "community_preferences"
CREATE INDEX "idx_community_preferences_deleted_at" ON "community_preferences" ("deleted_at");
-- Create "notification_preferences" table
CREATE TABLE "notification_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "enable_email_notification" boolean NOT NULL DEFAULT true,
  "enable_sms_notification" boolean NOT NULL DEFAULT false,
  "enable_push_notification" boolean NOT NULL DEFAULT false,
  "campaign_digest_enabled" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_notification_pref" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_notification_preferences_account_id" to table: "notification_preferences"
CREATE UNIQUE INDEX "idx_notification_preferences_account_id" ON "notification_preferences" ("account_id");
-- Create index "idx_notification_preferences_deleted_at" to table: "notification_preferences"
CREATE INDEX "idx_notification_preferences_deleted_at" ON "notification_preferences" ("deleted_at");
-- Create "oauth_identities" table
CREATE TABLE "oauth_identities" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "provider" character varying(100) NOT NULL,
  "provider_subject" character varying(255) NOT NULL,
  "provider_email" character varying(255) NULL,
  "access_token" text NULL,
  "refresh_token" text NULL,
  "token_expires_at" timestamptz NULL,
  "last_used_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_o_auth_identities" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_oauth_identities_account_id" to table: "oauth_identities"
CREATE INDEX "idx_oauth_identities_account_id" ON "oauth_identities" ("account_id");
-- Create index "idx_oauth_identities_deleted_at" to table: "oauth_identities"
CREATE INDEX "idx_oauth_identities_deleted_at" ON "oauth_identities" ("deleted_at");
-- Create index "idx_oauth_provider_subject" to table: "oauth_identities"
CREATE UNIQUE INDEX "idx_oauth_provider_subject" ON "oauth_identities" ("provider", "provider_subject");
-- Create "roles" table
CREATE TABLE "roles" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "code" character varying(100) NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "type" character varying(32) NOT NULL DEFAULT 'system',
  "is_system" boolean NOT NULL DEFAULT true,
  "is_mutable" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "idx_roles_code" to table: "roles"
CREATE UNIQUE INDEX "idx_roles_code" ON "roles" ("code");
-- Create index "idx_roles_deleted_at" to table: "roles"
CREATE INDEX "idx_roles_deleted_at" ON "roles" ("deleted_at");
-- Create "role_assignments" table
CREATE TABLE "role_assignments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "role_id" uuid NOT NULL,
  "assigned_by" uuid NOT NULL,
  "expires_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "revoke_reason" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_role_assignments" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_role_assignments_assigned_by_account" FOREIGN KEY ("assigned_by") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE RESTRICT,
  CONSTRAINT "fk_roles_role_assignments" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_role_assignments_account_role" to table: "role_assignments"
CREATE UNIQUE INDEX "idx_role_assignments_account_role" ON "role_assignments" ("account_id", "role_id");
-- Create index "idx_role_assignments_assigned_by" to table: "role_assignments"
CREATE INDEX "idx_role_assignments_assigned_by" ON "role_assignments" ("assigned_by");
-- Create index "idx_role_assignments_deleted_at" to table: "role_assignments"
CREATE INDEX "idx_role_assignments_deleted_at" ON "role_assignments" ("deleted_at");
-- Create index "idx_role_assignments_role_id" to table: "role_assignments"
CREATE INDEX "idx_role_assignments_role_id" ON "role_assignments" ("role_id");
-- Create "permissions" table
CREATE TABLE "permissions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "code" character varying(150) NOT NULL,
  "name" character varying(150) NOT NULL,
  "description" text NULL,
  "module" character varying(100) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_permissions_code" to table: "permissions"
CREATE UNIQUE INDEX "idx_permissions_code" ON "permissions" ("code");
-- Create index "idx_permissions_deleted_at" to table: "permissions"
CREATE INDEX "idx_permissions_deleted_at" ON "permissions" ("deleted_at");
-- Create index "idx_permissions_module" to table: "permissions"
CREATE INDEX "idx_permissions_module" ON "permissions" ("module");
-- Create "role_permissions" table
CREATE TABLE "role_permissions" (
  "role_id" uuid NOT NULL,
  "permission_id" uuid NOT NULL,
  CONSTRAINT "fk_permissions_role_permissions" FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON UPDATE CASCADE ON DELETE CASCADE,
  CONSTRAINT "fk_roles_role_permissions" FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_role_permissions_role_permission" to table: "role_permissions"
CREATE UNIQUE INDEX "idx_role_permissions_role_permission" ON "role_permissions" ("role_id", "permission_id");
-- Create "sessions" table
CREATE TABLE "sessions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "refresh_token_hash" character varying(255) NOT NULL,
  "user_agent" text NULL,
  "ip_address" character varying(64) NULL,
  "last_active_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamptz NOT NULL,
  "revoked_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_sessions" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_sessions_account_id" to table: "sessions"
CREATE INDEX "idx_sessions_account_id" ON "sessions" ("account_id");
-- Create index "idx_sessions_deleted_at" to table: "sessions"
CREATE INDEX "idx_sessions_deleted_at" ON "sessions" ("deleted_at");
-- Create index "idx_sessions_expires_at" to table: "sessions"
CREATE INDEX "idx_sessions_expires_at" ON "sessions" ("expires_at");
-- Create index "idx_sessions_refresh_token_hash" to table: "sessions"
CREATE UNIQUE INDEX "idx_sessions_refresh_token_hash" ON "sessions" ("refresh_token_hash");
-- Create "template_preferences" table
CREATE TABLE "template_preferences" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "default_template" character varying(100) NULL,
  "editor_mode" character varying(50) NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_accounts_template_preference" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_template_preferences_account_id" to table: "template_preferences"
CREATE UNIQUE INDEX "idx_template_preferences_account_id" ON "template_preferences" ("account_id");
-- Create index "idx_template_preferences_deleted_at" to table: "template_preferences"
CREATE INDEX "idx_template_preferences_deleted_at" ON "template_preferences" ("deleted_at");
