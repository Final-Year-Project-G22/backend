-- Modify "guide_steps" table
ALTER TABLE "guide_steps" ALTER COLUMN "required_documents" DROP NOT NULL, ALTER COLUMN "external_links" DROP NOT NULL;
