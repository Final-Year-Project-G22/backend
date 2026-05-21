-- Seed compliance type localized labels (en + am)
-- Apply after migration add_scheduled_alerts_compliance_and_guide_compliance_type

INSERT INTO compliance_type_localizations (id, compliance_type, locale, label)
VALUES
  (gen_random_uuid(), 'tax_registration',      'en', 'Tax Registration'),
  (gen_random_uuid(), 'tax_registration',      'am', 'የግብር ምዝገባ'),
  (gen_random_uuid(), 'trade_license',         'en', 'Trade License'),
  (gen_random_uuid(), 'trade_license',         'am', 'የንግድ ፍቃድ'),
  (gen_random_uuid(), 'business_registration', 'en', 'Business Registration'),
  (gen_random_uuid(), 'business_registration', 'am', 'የንግድ ምዝገባ')
ON CONFLICT (compliance_type, locale) DO UPDATE SET
  label      = EXCLUDED.label,
  updated_at = CURRENT_TIMESTAMP;
