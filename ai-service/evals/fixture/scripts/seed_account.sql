-- ============================================================================
-- Evaluation Fixture: seeded core-backend account state (FIN-76)
-- Fixture version: 1.0.0  (see ai-service/evals/fixture/manifest.json)
--
-- Creates the fixture account "eval-msme-01" with a known business profile,
-- compliance entries, a fixture guide with steps, and guide progress.
-- Uses fixed UUIDs so the state is reproducible across environments.
-- No live/demo data: all rows are authored fixture content.
--
-- Run against the core-backend database (e.g. psql $CORE_DB < seed_account.sql)
-- ============================================================================

-- --- Fixture user + account -------------------------------------------------
INSERT INTO users (id, first_name, last_name)
VALUES ('10000000-0000-4000-8000-000000000001', 'Eval', 'Msme')
ON CONFLICT (id) DO NOTHING;

INSERT INTO accounts (
    id, user_id, email, email_normalized, password_hash,
    email_verified, phone_verified, status
)
VALUES (
    '10000000-0000-4000-8000-000000000002',
    '10000000-0000-4000-8000-000000000001',
    'eval-msme-01@fixture.local',
    'EVAL-MSME-01@FIXTURE.LOCAL',
    '$2y$10$2mi0Jgaggkgp1nKx3Y56DeqX3lnbk31WK2eQPNxfIYPH1TRf26G3q',
    true, false, 'active'
)
ON CONFLICT (id) DO NOTHING;

-- --- Business profile --------------------------------------------------------
-- sector: crop-farming (looked up by slug, environment-independent)
INSERT INTO business_profiles (
    id, account_id, company_name, company_email, company_phone_number,
    physical_address, description, region, stage, sector_id,
    registration_number, registration_date, tax_identification_number,
    trade_license_number
)
SELECT
    '10000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000002',
    'Selam Coffee Export PLC',
    'contact@selamcoffee.fixture.local',
    '+251911000001',
    'Bole Sub-city, Addis Ababa',
    'Fixture MSME: coffee export business used for AI evaluation.',
    'OROMIA',
    'OPERATIONAL',
    s.id,
    'BR-2026-00777',
    DATE '2025-06-15',
    'TIN-100-2026-0001',
    'TL-AA-2026-0042'
FROM sectors s
WHERE s.slug = 'crop-farming'
ON CONFLICT (id) DO NOTHING;

-- profile tags: plc, op-exporter, has-employees, tax-vat (looked up by slug)
INSERT INTO business_profile_tags (business_profile_id, tag_id)
SELECT
    '10000000-0000-4000-8000-000000000003',
    t.id
FROM tags t
WHERE t.slug IN ('plc', 'op-exporter', 'has-employees', 'tax-vat')
ON CONFLICT (business_profile_id, tag_id) DO NOTHING;

-- --- Compliance entries -------------------------------------------------------
-- trade_license: due for renewal (expires 2027-01-10, reminder window 30 days)
INSERT INTO compliance_entries (
    id, business_profile_id, account_id, compliance_type, reference_number,
    issued_date, expiry_date, reminder_days_before, source, status
)
VALUES (
    '10000000-0000-4000-8000-000000000004',
    '10000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000002',
    'trade_license',
    'TL-AA-2026-0042',
    DATE '2026-01-10',
    TIMESTAMPTZ '2027-01-10 23:59:59+00',
    30, 'manual', 'active'
);

-- tax_registration: TIN, long validity
INSERT INTO compliance_entries (
    id, business_profile_id, account_id, compliance_type, reference_number,
    issued_date, expiry_date, reminder_days_before, source, status
)
VALUES (
    '10000000-0000-4000-8000-000000000005',
    '10000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000002',
    'tax_registration',
    'TIN-100-2026-0001',
    DATE '2025-06-15',
    TIMESTAMPTZ '2030-06-15 23:59:59+00',
    45, 'manual', 'active'
);

-- business_registration: long validity
INSERT INTO compliance_entries (
    id, business_profile_id, account_id, compliance_type, reference_number,
    issued_date, expiry_date, reminder_days_before, source, status
)
VALUES (
    '10000000-0000-4000-8000-000000000006',
    '10000000-0000-4000-8000-000000000003',
    '10000000-0000-4000-8000-000000000002',
    'business_registration',
    'BR-2026-00777',
    DATE '2025-06-15',
    TIMESTAMPTZ '2031-06-15 23:59:59+00',
    60, 'manual', 'active'
);

-- --- Fixture guide: "Business formalization" with steps -----------------------
INSERT INTO guides (id, slug, icon, sort_order, sector_ids, tag_ids)
SELECT
    '20000000-0000-4000-8000-000000000001',
    'fixture-business-formalization',
    'briefcase',
    1,
    ARRAY[s.id],
    ARRAY[t.id]
FROM sectors s, tags t
WHERE s.slug = 'crop-farming' AND t.slug = 'plc'
ON CONFLICT (id) DO NOTHING;

-- guide translations
INSERT INTO guide_translations (id, guide_id, language, name, description)
VALUES
    ('20000000-0000-4000-8000-000000000011', '20000000-0000-4000-8000-000000000001', 'en',
     'Business Formalization', 'Steps to register and license your coffee export business'),
    ('20000000-0000-4000-8000-000000000012', '20000000-0000-4000-8000-000000000001', 'am',
     'የንግድ ሕጋዊነት', 'የቡና ኤክስፖርት ንግድዎን ለመመዝገብ እና ፈቃድ ለማግኘት የሚያስፈልጉ ደረጃዎች')
ON CONFLICT (id) DO NOTHING;

-- guide steps (informational -> action -> document submission -> verification)
INSERT INTO guide_steps (id, guide_id, slug, step_type, sort_order, estimated_time, difficulty_level, is_optional, fee_estimate, version, effective_date, compliance_type)
VALUES
    ('21000000-0000-4000-8000-000000000001', '20000000-0000-4000-8000-000000000001', 'register-company', 'ACTION_REQUIRED', 1, 60, 2, false, 0, 1, DATE '2026-01-01', 'business_registration'),
    ('21000000-0000-4000-8000-000000000002', '20000000-0000-4000-8000-000000000001', 'obtain-tin', 'DOCUMENT_SUBMISSION', 2, 30, 1, false, 0, 1, DATE '2026-01-01', 'tax_registration'),
    ('21000000-0000-4000-8000-000000000003', '20000000-0000-4000-8000-000000000001', 'renew-trade-license', 'ACTION_REQUIRED', 3, 45, 2, false, 500, 1, DATE '2026-01-01', 'trade_license'),
    ('21000000-0000-4000-8000-000000000004', '20000000-0000-4000-8000-000000000001', 'compliance-check', 'VERIFICATION', 4, 20, 1, false, 0, 1, DATE '2026-01-01', NULL)
ON CONFLICT (id) DO NOTHING;

-- step translations
INSERT INTO guide_step_translations (id, guide_step_id, language, title, description, required_documents, detailed_content)
VALUES
    ('21100000-0000-4000-8000-000000000001', '21000000-0000-4000-8000-000000000001', 'en', 'Register your company', 'Register the PLC with the trade ministry', '["memorandum", "articles"]', '{"what": "Submit registration documents"}'),
    ('21100000-0000-4000-8000-000000000002', '21000000-0000-4000-8000-000000000001', 'am', 'ኩባንያዎን ያስመዝግቡ', 'ኩባንያዎን በንግድ ሚኒስቴር ያስመዝግቡ', '["ማስታወቂያ", "አሠራር ደንብ"]', '{"what": "የምዝገባ ሰነዶችን ያቅርቡ"}'),
    ('21100000-0000-4000-8000-000000000003', '21000000-0000-4000-8000-000000000002', 'en', 'Obtain your TIN', 'Get a taxpayer identification number from the tax authority', '[]', '{"what": "Apply for TIN"}'),
    ('21100000-0000-4000-8000-000000000004', '21000000-0000-4000-8000-000000000002', 'am', 'TIN ያግኙ', 'ከግብር ባለሥልጣን የታክስ ከፋይ መለያ ቁጥር ያግኙ', '[]', '{"what": "ለTIN ያመልክቱ"}'),
    ('21100000-0000-4000-8000-000000000005', '21000000-0000-4000-8000-000000000003', 'en', 'Renew your trade license', 'Renew the annual trade license before expiry', '["trade_license"]', '{"what": "Pay renewal fee"}'),
    ('21100000-0000-4000-8000-000000000006', '21000000-0000-4000-8000-000000000003', 'am', 'የንግድ ፈቃድዎን ያድሱ', 'ዓመታዊ የንግድ ፈቃድዎን ከማብቃቱ በፊት ያድሱ', '["የንግድ ፈቃድ"]', '{"what": "የእድሳት ክፍያ ይክፈሉ"}'),
    ('21100000-0000-4000-8000-000000000007', '21000000-0000-4000-8000-000000000004', 'en', 'Run a compliance check', 'Verify all registrations and licenses are current', '[]', '{"what": "Review compliance status"}'),
    ('21100000-0000-4000-8000-000000000008', '21000000-0000-4000-8000-000000000004', 'am', 'የተገዢነት ምርመራ ያድርጉ', 'ሁሉም ምዝገባዎች እና ፈቃዶች የተዘመኑ መሆናቸውን ያረጋግጡ', '[]', '{"what": "የተገዢነት ሁኔታን ይገምግሙ"}')
ON CONFLICT (id) DO NOTHING;

-- --- Guide progress (steps 1-2 completed, step 3 in progress, step 4 locked) --
INSERT INTO user_guide_progresses (
    id, account_id, user_id, step_id, status, started_at, completed_at,
    time_spent, notes, uploaded_documents, last_accessed_at, version
)
VALUES
    ('30000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', '21000000-0000-4000-8000-000000000001', 'COMPLETED', TIMESTAMPTZ '2026-01-15 10:00:00+00', TIMESTAMPTZ '2026-01-15 11:00:00+00', 60, 'Registered with ministry', '[]', TIMESTAMPTZ '2026-01-15 11:00:00+00', 1),
    ('30000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', '21000000-0000-4000-8000-000000000002', 'COMPLETED', TIMESTAMPTZ '2026-01-16 09:00:00+00', TIMESTAMPTZ '2026-01-16 09:30:00+00', 30, 'TIN obtained', '[]', TIMESTAMPTZ '2026-01-16 09:30:00+00', 1),
    ('30000000-0000-4000-8000-000000000003', '10000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', '21000000-0000-4000-8000-000000000003', 'IN_PROGRESS', TIMESTAMPTZ '2026-01-20 09:00:00+00', NULL, 20, 'Renewal pending', '[]', TIMESTAMPTZ '2026-01-20 09:20:00+00', 1),
    ('30000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', '21000000-0000-4000-8000-000000000004', 'LOCKED', NULL, NULL, NULL, NULL, '[]', NULL, 1)
ON CONFLICT (id) DO NOTHING;

-- --- Preferences --------------------------------------------------------------
INSERT INTO account_preferences (account_id, language, timezone)
VALUES ('10000000-0000-4000-8000-000000000002', 'en', 'Africa/Addis_Ababa')
ON CONFLICT (account_id) DO NOTHING;

INSERT INTO ai_preferences (account_id, default_model, response_style, temperature, allow_data_retention)
VALUES ('10000000-0000-4000-8000-000000000002', 'default', 'concise', 0.2, true)
ON CONFLICT (account_id) DO NOTHING;
