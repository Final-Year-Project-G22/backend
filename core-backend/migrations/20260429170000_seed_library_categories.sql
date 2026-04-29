-- Seed initial library categories with English translations.
-- These are root-level categories (parent_category_id IS NULL).

INSERT INTO library_categories (name, slug, icon, sort_order, is_active)
VALUES
  ('Business Plans',        'business-plans',        'briefcase',   1, true),
  ('Invoices',              'invoices',              'receipt',     2, true),
  ('Record Keeping',        'record-keeping',        'archive',     3, true),
  ('Financial Statements',  'financial-statements',  'chart',       4, true),
  ('Contracts & Agreements','contracts-agreements',  'file-text',   5, true),
  ('Marketing Materials',   'marketing-materials',   'megaphone',   6, true);
