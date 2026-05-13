-- Modify "guide_step_translations" table
ALTER TABLE "guide_step_translations" ADD COLUMN "required_documents" jsonb NULL DEFAULT '[]';
-- Modify "guide_steps" table
ALTER TABLE "guide_steps" DROP COLUMN "required_documents";
