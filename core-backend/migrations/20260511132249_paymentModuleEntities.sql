-- Create "payments" table
CREATE TABLE "payments" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "subscription_id" uuid NULL,
  "tx_ref" character varying(255) NOT NULL,
  "chapa_ref" character varying(255) NULL,
  "amount" bigint NOT NULL,
  "currency" character varying(3) NOT NULL DEFAULT 'ETB',
  "plan_name" character varying(50) NOT NULL,
  "plan_period" character varying(20) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "payment_method" character varying(50) NULL,
  "checked_out_at" timestamptz NULL,
  "verified_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "metadata" jsonb NULL DEFAULT '{}',
  PRIMARY KEY ("id")
);
-- Create index "idx_payments_account_id" to table: "payments"
CREATE INDEX "idx_payments_account_id" ON "payments" ("account_id");
-- Create index "idx_payments_account_status" to table: "payments"
CREATE INDEX "idx_payments_account_status" ON "payments" ("account_id", "status");
-- Create index "idx_payments_chapa_ref" to table: "payments"
CREATE INDEX "idx_payments_chapa_ref" ON "payments" ("chapa_ref");
-- Create index "idx_payments_deleted_at" to table: "payments"
CREATE INDEX "idx_payments_deleted_at" ON "payments" ("deleted_at");
-- Create index "idx_payments_status" to table: "payments"
CREATE INDEX "idx_payments_status" ON "payments" ("status");
-- Create index "idx_payments_subscription_id" to table: "payments"
CREATE INDEX "idx_payments_subscription_id" ON "payments" ("subscription_id");
-- Create index "idx_payments_tx_ref" to table: "payments"
CREATE UNIQUE INDEX "idx_payments_tx_ref" ON "payments" ("tx_ref");
-- Create "plans" table
CREATE TABLE "plans" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "name" character varying(50) NOT NULL,
  "period" character varying(20) NOT NULL,
  "amount" bigint NOT NULL,
  "currency" character varying(3) NOT NULL DEFAULT 'ETB',
  "is_active" boolean NOT NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_plans_deleted_at" to table: "plans"
CREATE INDEX "idx_plans_deleted_at" ON "plans" ("deleted_at");
-- Create index "idx_plans_name_period" to table: "plans"
CREATE UNIQUE INDEX "idx_plans_name_period" ON "plans" ("name", "period");
-- Create "subscriptions" table
CREATE TABLE "subscriptions" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "deleted_at" timestamptz NULL,
  "account_id" uuid NOT NULL,
  "plan_name" character varying(50) NOT NULL,
  "plan_period" character varying(20) NOT NULL,
  "amount" bigint NOT NULL,
  "currency" character varying(3) NOT NULL DEFAULT 'ETB',
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "current_period_start" timestamptz NOT NULL,
  "current_period_end" timestamptz NOT NULL,
  "cancelled_at" timestamptz NULL,
  "renewal_count" bigint NOT NULL DEFAULT 0,
  PRIMARY KEY ("id")
);
-- Create index "idx_subscriptions_account" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_account" ON "subscriptions" ("account_id");
-- Create index "idx_subscriptions_deleted_at" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_deleted_at" ON "subscriptions" ("deleted_at");
-- Create index "idx_subscriptions_period_end" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_period_end" ON "subscriptions" ("current_period_end");
-- Create index "idx_subscriptions_status" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_status" ON "subscriptions" ("status");
