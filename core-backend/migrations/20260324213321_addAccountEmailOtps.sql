-- Create "account_email_otps" table
CREATE TABLE "account_email_otps" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "code_hash" character varying(255) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "consumed_at" timestamptz NULL,
  "attempt_count" bigint NOT NULL DEFAULT 0,
  "resend_count" bigint NOT NULL DEFAULT 0,
  "last_sent_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_account_email_otps_account" FOREIGN KEY ("account_id") REFERENCES "accounts" ("id") ON UPDATE CASCADE ON DELETE CASCADE
);
-- Create index "idx_account_email_otps_account_id" to table: "account_email_otps"
CREATE INDEX "idx_account_email_otps_account_id" ON "account_email_otps" ("account_id");
-- Create index "idx_account_email_otps_deleted_at" to table: "account_email_otps"
CREATE INDEX "idx_account_email_otps_deleted_at" ON "account_email_otps" ("deleted_at");
-- Create index "idx_account_email_otps_expires_at" to table: "account_email_otps"
CREATE INDEX "idx_account_email_otps_expires_at" ON "account_email_otps" ("expires_at");
