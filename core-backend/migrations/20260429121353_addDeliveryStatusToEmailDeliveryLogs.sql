-- Modify "email_delivery_logs" table
ALTER TABLE "email_delivery_logs" ADD COLUMN "delivery_status" character varying(20) NOT NULL DEFAULT 'sent';
