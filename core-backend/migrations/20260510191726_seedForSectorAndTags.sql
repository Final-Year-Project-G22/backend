-- =================================================================================
-- ADISU SERATEGNA: UNIFIED BUSINESS TAXONOMY SEEDER
-- =================================================================================

-- ---------------------------------------------------------------------------------
-- 1. SECTORS (Pillar 1: Industry Verticals)
-- ---------------------------------------------------------------------------------

-- A. Insert Root Sectors
INSERT INTO sectors (slug, parent_id, ancestor_ids)
VALUES
  ('trade',         NULL, NULL),
  ('manufacturing', NULL, NULL),
  ('services',      NULL, NULL),
  ('agriculture',   NULL, NULL),
  ('construction',  NULL, NULL);

-- B. Insert Child Sectors (parent resolved by slug)
INSERT INTO sectors (slug, parent_id, ancestor_ids)
VALUES
  -- Trade
  ('retail',             (SELECT id FROM sectors WHERE slug = 'trade'), NULL),
  ('wholesale',          (SELECT id FROM sectors WHERE slug = 'trade'), NULL),
  -- Manufacturing
  ('food-beverage',      (SELECT id FROM sectors WHERE slug = 'manufacturing'), NULL),
  ('textiles-apparel',   (SELECT id FROM sectors WHERE slug = 'manufacturing'), NULL),
  ('leather',            (SELECT id FROM sectors WHERE slug = 'manufacturing'), NULL),
  ('wood-metal',         (SELECT id FROM sectors WHERE slug = 'manufacturing'), NULL),
  -- Services
  ('it-tech',            (SELECT id FROM sectors WHERE slug = 'services'), NULL),
  ('hospitality',        (SELECT id FROM sectors WHERE slug = 'services'), NULL),
  ('consulting',         (SELECT id FROM sectors WHERE slug = 'services'), NULL),
  ('transport-logistics',(SELECT id FROM sectors WHERE slug = 'services'), NULL),
  -- Agriculture
  ('crop-farming',       (SELECT id FROM sectors WHERE slug = 'agriculture'), NULL),
  ('livestock-poultry',  (SELECT id FROM sectors WHERE slug = 'agriculture'), NULL),
  -- Construction
  ('contracting',        (SELECT id FROM sectors WHERE slug = 'construction'), NULL),
  ('construction-mat',   (SELECT id FROM sectors WHERE slug = 'construction'), NULL);

-- C. Update ancestor_ids for root sectors (self only)
UPDATE sectors
SET ancestor_ids = ARRAY[id]
WHERE parent_id IS NULL;

-- D. Update ancestor_ids for child sectors (parent ancestors + self)
UPDATE sectors child
SET ancestor_ids = parent.ancestor_ids || child.id
FROM sectors parent
WHERE child.parent_id = parent.id;

-- E. Seed Sector Translations (English & Amharic)
INSERT INTO sector_translations (sector_id, language, name, description)
SELECT id, 'en', 'Trade', 'Retail and wholesale distribution of goods' FROM sectors WHERE slug = 'trade' UNION ALL
SELECT id, 'am', 'ንግድ', 'የእቃ መሸጥ እና ማከፋፈል ስራዎች' FROM sectors WHERE slug = 'trade' UNION ALL

SELECT id, 'en', 'Manufacturing', 'Industrial production and processing' FROM sectors WHERE slug = 'manufacturing' UNION ALL
SELECT id, 'am', 'ማምረቻ', 'የኢንዱስትሪ ምርት እና ማቀነባበር' FROM sectors WHERE slug = 'manufacturing' UNION ALL

SELECT id, 'en', 'Services', 'Service-based businesses and consulting' FROM sectors WHERE slug = 'services' UNION ALL
SELECT id, 'am', 'አገልግሎት', 'የአገልግሎት ሰጪ እና አማካሪ ድርጅቶች' FROM sectors WHERE slug = 'services' UNION ALL

SELECT id, 'en', 'Agriculture', 'Farming, livestock, and agro-processing' FROM sectors WHERE slug = 'agriculture' UNION ALL
SELECT id, 'am', 'ግብርና', 'እርሻ፣ እንስሳት እርባታ እና ግብርና ነክ ምርቶች' FROM sectors WHERE slug = 'agriculture' UNION ALL

SELECT id, 'en', 'Construction', 'Building, contracting, and materials' FROM sectors WHERE slug = 'construction' UNION ALL
SELECT id, 'am', 'ኮንስትራክሽን', 'የህንፃ፣ ተቋራጭነት እና የግንባታ እቃዎች' FROM sectors WHERE slug = 'construction';

-- (Note: You can easily add the child translations following the exact pattern above)


-- ---------------------------------------------------------------------------------
-- 2. TAGS (Pillar 2: Operations, Taxes, and Legal Status)
-- ---------------------------------------------------------------------------------

-- A. Seed Canonical Tags with Groups
INSERT INTO tags (slug, "group", is_multi_select)
VALUES
  -- LEGAL_STRUCTURE (Single Select - dictates base registration rules)
  ('sole-proprietor',     'LEGAL_STRUCTURE', false),
  ('plc',                 'LEGAL_STRUCTURE', false),
  ('share-company',       'LEGAL_STRUCTURE', false),
  ('partnership',         'LEGAL_STRUCTURE', false),
  ('cooperative',         'LEGAL_STRUCTURE', false),

  -- TAX_STATUS (Single Select - dictates recurring compliance alerts)
  ('tax-vat',             'TAX_STATUS', false),
  ('tax-tot',             'TAX_STATUS', false),
  ('tax-excise',          'TAX_STATUS', false),
  ('tax-exempt',          'TAX_STATUS', false),

  -- GENERAL_OPERATIONS (Multi Select - dictates cross-ministry compliance)
  ('op-importer',         'GENERAL_OPERATIONS', true),
  ('op-exporter',         'GENERAL_OPERATIONS', true),
  ('op-tender',           'GENERAL_OPERATIONS', true),
  ('op-food-handling',    'GENERAL_OPERATIONS', true),
  ('op-vehicles',         'GENERAL_OPERATIONS', true),
  ('op-hazardous',        'GENERAL_OPERATIONS', true),
  ('op-ecommerce',        'GENERAL_OPERATIONS', true),
  ('op-home-based',       'GENERAL_OPERATIONS', true),
  ('op-creates-ip',       'GENERAL_OPERATIONS', true),

  -- EMPLOYMENT (Single Select - dictates payroll/pension laws)
  ('has-employees',       'EMPLOYMENT', false),
  ('no-employees',        'EMPLOYMENT', false),

  -- DEMOGRAPHICS (Multi Select - opens specific AI contexts & grants)
  ('demo-women-owned',    'DEMOGRAPHICS', true),
  ('demo-youth',          'DEMOGRAPHICS', true),
  ('demo-investor',       'DEMOGRAPHICS', true);


-- B. Seed Tag Translations (English & Amharic)
INSERT INTO tag_translations (tag_id, language, name, description)
-- Legal Structure
SELECT id, 'en', 'Sole Proprietor', 'Business owned and run by one individual' FROM tags WHERE slug = 'sole-proprietor' UNION ALL
SELECT id, 'am', 'የግል ማህበር', 'በአንድ ግለሰብ የተመሰረተ ንግድ' FROM tags WHERE slug = 'sole-proprietor' UNION ALL

SELECT id, 'en', 'Private Limited Company (PLC)', 'Company with limited liability' FROM tags WHERE slug = 'plc' UNION ALL
SELECT id, 'am', 'ኃላፊነቱ የተወሰነ የግል ማህበር', 'በአክሲዮን የተወሰነ ኃላፊነት ያለው' FROM tags WHERE slug = 'plc' UNION ALL

-- Tax Status
SELECT id, 'en', 'VAT Payer', 'Registered for Value Added Tax (>2M ETB)' FROM tags WHERE slug = 'tax-vat' UNION ALL
SELECT id, 'am', 'የተጨማሪ እሴት ታክስ (ተ.እ.ታ)', 'ከ 2 ሚሊዮን ብር በላይ ገቢ ያላቸው' FROM tags WHERE slug = 'tax-vat' UNION ALL

SELECT id, 'en', 'TOT Payer', 'Registered for Turnover Tax (<2M ETB)' FROM tags WHERE slug = 'tax-tot' UNION ALL
SELECT id, 'am', 'የሽያጭ ታክስ', 'ከ 2 ሚሊዮን ብር በታች ገቢ ያላቸው' FROM tags WHERE slug = 'tax-tot' UNION ALL

-- General Operations
SELECT id, 'en', 'Importer', 'Business that imports goods into Ethiopia' FROM tags WHERE slug = 'op-importer' UNION ALL
SELECT id, 'am', 'አስመጪ', 'እቃዎችን ወደ ሀገር ውስጥ የሚያስገባ' FROM tags WHERE slug = 'op-importer' UNION ALL

SELECT id, 'en', 'Exporter', 'Business that exports goods from Ethiopia' FROM tags WHERE slug = 'op-exporter' UNION ALL
SELECT id, 'am', 'ላኪ', 'እቃዎችን ወደ ውጪ የሚልክ' FROM tags WHERE slug = 'op-exporter' UNION ALL

SELECT id, 'en', 'Food & Beverage Handling', 'Produces or serves food/drinks' FROM tags WHERE slug = 'op-food-handling' UNION ALL
SELECT id, 'am', 'ምግብና መጠጥ ነክ', 'ምግብ ወይም መጠጥ የሚያዘጋጅ' FROM tags WHERE slug = 'op-food-handling' UNION ALL

-- Employment
SELECT id, 'en', 'Has Employees', 'Employs 1 or more salaried workers' FROM tags WHERE slug = 'has-employees' UNION ALL
SELECT id, 'am', 'ሰራተኛ ያለው', '1 እና ከዚያ በላይ ሰራተኛ ያለው' FROM tags WHERE slug = 'has-employees' UNION ALL

SELECT id, 'en', 'No Employees (Solopreneur)', 'No official salaried workers' FROM tags WHERE slug = 'no-employees' UNION ALL
SELECT id, 'am', 'ሰራተኛ የሌለው', 'በግል የሚሰራ (ሰራተኛ የሌለው)' FROM tags WHERE slug = 'no-employees' UNION ALL

-- Demographics
SELECT id, 'en', 'Women-Owned', 'Enterprise founded/owned by a woman' FROM tags WHERE slug = 'demo-women-owned' UNION ALL
SELECT id, 'am', 'በሴቶች የተመሰረተ', 'በሴቶች ባለቤትነት የሚተዳደር' FROM tags WHERE slug = 'demo-women-owned' UNION ALL

SELECT id, 'en', 'Youth Enterprise', 'Owned by youth (aged 18-35)' FROM tags WHERE slug = 'demo-youth' UNION ALL
SELECT id, 'am', 'የወጣቶች ኢንተርፕራይዝ', 'በወጣቶች የተመሰረተ (18-35)' FROM tags WHERE slug = 'demo-youth';

-- (This successfully seeds the core architecture needed for the Flutter UI and AI Context)
