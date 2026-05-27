-- Seed Amharic translations for library categories.

INSERT INTO library_category_translations (library_category_id, language, name, description)
SELECT id, 'am', 'የንግድ ዕቅዶች', 'የንግድ መነሻ እና ልማት ዕቅዶች' FROM library_categories WHERE slug = 'business-plans' UNION ALL
SELECT id, 'am', 'ደንበኞች ሂሳብ', 'የደንበኛ ሂሳብ እና የፋይናንስ ማቅረቢያ' FROM library_categories WHERE slug = 'invoices' UNION ALL

SELECT id, 'am', 'የሂሳብ መዝገብ', 'የንግድ ሂሳብ እና የፋይናንስ መዝገቦች' FROM library_categories WHERE slug = 'record-keeping' UNION ALL

SELECT id, 'am', 'የፋይናንስ መግለጫዎች', 'የቂነት መግለጫ እና የፋይናንስ ሪፖርቶች' FROM library_categories WHERE slug = 'financial-statements' UNION ALL

SELECT id, 'am', 'ውሎች እና ስምምነቶች', 'የንግድ ውሎች እና የስምምነት ሰነዶች' FROM library_categories WHERE slug = 'contracts-agreements' UNION ALL

SELECT id, 'am', 'የገበያ ማሰሪያ ሰነዶች', 'የንግድ ማሰሪያ እና የማስተዋይ ሰነዶች' FROM library_categories WHERE slug = 'marketing-materials';
