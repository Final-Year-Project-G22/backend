-- Modify "account_email_otps" table
ALTER TABLE "account_email_otps" ADD COLUMN "purpose" character varying(32) NOT NULL DEFAULT 'email_verification';
