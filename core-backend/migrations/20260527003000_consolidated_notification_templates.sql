-- Consolidated migration:
--   1. Update email HTML (no links) + add push for 6 templates
--   2. Seed admin_provisioned template
--   3. Add Amharic translations for all 13 templates missing them
--   4. account_verification uses {{otpCode}} (not {{verificationMessage}})

----------------------------------------------------------------------
-- PART 1: Update default_content with styled email HTML (no links) + push
----------------------------------------------------------------------

-- 1a. system_announcement (email only)
UPDATE notification_templates
SET default_content = jsonb_set(
    default_content,
    '{email}',
    '{"subject":"System Announcement: {{message}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">System Announcement</h2><p style=\"font-size:15px;line-height:1.7;\">{{message}}</p></div>"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'system_announcement';

-- 1b. policy_update (email only)
UPDATE notification_templates
SET default_content = jsonb_set(
    default_content,
    '{email}',
    '{"subject":"Important: Policy Update","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#d97706;\">Policy Update</h2><p style=\"font-size:15px;line-height:1.7;\">{{summary}}</p></div>"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'policy_update';

-- 1c. welcome_message (email only)
UPDATE notification_templates
SET default_content = jsonb_set(
    default_content,
    '{email}',
    '{"subject":"Welcome to {{platformName}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#16a34a;\">Welcome to {{platformName}}</h2><p style=\"font-size:15px;line-height:1.7;\">Hi {{accountName}},</p><p style=\"font-size:15px;line-height:1.7;\">We''re excited to have you on board. Start exploring guides and discover everything available to you.</p></div>"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'welcome_message';

-- 1d. guide_deadline (email + push)
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"Deadline Approaching: {{stepTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">Deadline Approaching</h2><p style=\"font-size:15px;line-height:1.7;\">The step <strong>{{stepTitle}}</strong> in <strong>{{guideName}}</strong> is due by <strong>{{deadlineDate}}</strong>.</p></div>"}'::jsonb,
        true
    ),
    '{push}',
    '{"title":"Deadline Approaching","body":"{{stepTitle}} in {{guideName}} is due by {{deadlineDate}}"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'guide_deadline';

-- 1e. ai_quota_limit (email + push)
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"AI Usage Alert: {{percentUsed}}% Used","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#7c3aed;\">AI Usage Alert</h2><p style=\"font-size:15px;line-height:1.7;\">You have used <strong>{{percentUsed}}%</strong> of your AI quota.</p><p style=\"font-size:15px;line-height:1.7;\">{{message}}</p></div>"}'::jsonb,
        true
    ),
    '{push}',
    '{"title":"AI Quota {{status}}","body":"You have {{percentUsed}}% of your AI quota. {{message}}"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'ai_quota_limit';

-- 1f. ai_response_ready (add push, in-app already exists, no email)
UPDATE notification_templates
SET default_content = jsonb_set(
    default_content,
    '{push}',
    '{"title":"AI Response Ready","body":"Your AI request \"{{queryPreview}}\" has completed"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'ai_response_ready';

-- 1g. account_alert (email + push)
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"Security Alert: {{alertTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">Security Alert</h2><div style=\"background:#fef2f2;border-left:4px solid #dc2626;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:15px;\"><strong>{{alertTitle}}</strong></p><p style=\"margin:12px 0 0;font-size:15px;line-height:1.7;\">{{alertMessage}}</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;color:#4b5563;\">If you did not perform this action, secure your account immediately.</p></div>"}'::jsonb,
        true
    ),
    '{push}',
    '{"title":"Security Alert","body":"{{alertMessage}}"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'account_alert';

-- 1h. account_alert_critical (email + push)
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"[URGENT] Security Alert: {{alertTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#b91c1c;\">URGENT Security Alert</h2><div style=\"background:#fef2f2;border-left:4px solid #b91c1c;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:15px;\"><strong>{{alertTitle}}</strong></p><p style=\"margin:12px 0 0;font-size:15px;line-height:1.7;\">{{alertMessage}}</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;color:#4b5563;\">If you did not perform this action, secure your account immediately.</p></div>"}'::jsonb,
        true
    ),
    '{push}',
    '{"title":"Security Alert","body":"{{alertTitle}}: {{alertMessage}}"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'account_alert_critical';

-- 1i. account_verification (email only, uses {{otpCode}} instead of {{verificationMessage}})
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"Verify your {{platformName}} account","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">Verify Your Account</h2><p style=\"font-size:15px;line-height:1.7;\">Your verification code is <strong>{{otpCode}}</strong>. Please enter this code to verify your email.</p><p style=\"margin-top:20px;font-size:13px;color:#6b7280;\">This verification link expires in {{expiryMinutes}} minutes.</p></div>"}'::jsonb,
        true
    ),
    '{inapp}',
    '{"title":"Verification Required","body":"Your verification code is {{otpCode}}. Please enter this code to verify your email.","actionUrl":"/account/verify"}'::jsonb,
    true
),
variables_schema = '{"required":["platformName","otpCode","expiryMinutes"]}'::jsonb,
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'account_verification';

-- 1j. password_reset (email only)
UPDATE notification_templates
SET default_content = jsonb_set(
    default_content,
    '{email}',
    '{"subject":"Reset your {{platformName}} password","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">Password Reset Request</h2><p style=\"font-size:15px;line-height:1.7;\">We received a request to reset your {{platformName}} account password.</p><p style=\"margin-top:20px;font-size:13px;color:#6b7280;\">This reset link expires in {{expiryMinutes}} minutes.</p><p style=\"font-size:13px;color:#6b7280;\">If you did not request this, you can safely ignore this email.</p></div>"}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'password_reset';

-- 1k. payment_confirmation (email + push)
UPDATE notification_templates
SET default_content = jsonb_set(
    jsonb_set(default_content, '{email}',
        '{"subject":"Payment Confirmed: {{amount}} {{currency}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#16a34a;\">Payment Confirmed</h2><div style=\"background:#f0fdf4;border-left:4px solid #16a34a;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:16px;\"><strong>{{amount}} {{currency}}</strong></p><p style=\"margin:8px 0 0;font-size:14px;color:#374151;\">Your payment has been successfully processed.</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;\"><strong>Reference:</strong> {{referenceId}}<br/><strong>Date:</strong> {{paymentDate}}</p></div>"}'::jsonb,
        true
    ),
    '{push}',
    '{"title":"Payment Confirmed","body":"Your payment of {{amount}} {{currency}} has been confirmed."}'::jsonb,
    true
),
updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'payment_confirmation';

----------------------------------------------------------------------
-- PART 2: Seed admin_provisioned notification template
----------------------------------------------------------------------

INSERT INTO notification_templates (id, name, description, notification_type, template_group, priority, is_system_managed, default_content, variables_schema, default_ttl)
VALUES (
  gen_random_uuid(),
  'Admin Provisioned',
  'Email sent to newly created admin accounts with auto-generated password.',
  'admin_provisioned', 'system', 2, true,
  '{"email":{"subject":"Your {{platformName}} Admin Account Credentials","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">Admin Account Created</h2><p style=\"font-size:15px;line-height:1.7;\">Hi {{accountName}},</p><p style=\"font-size:15px;line-height:1.7;\">An admin account has been created for you on <strong>{{platformName}}</strong>.</p><div style=\"background:#f0f5ff;border-left:4px solid #2563eb;padding:16px;border-radius:8px;margin-top:16px;\"><p style=\"margin:0;font-size:15px;\"><strong>Email:</strong> {{email}}</p><p style=\"margin:12px 0 0;font-size:15px;\"><strong>Password:</strong> {{password}}</p></div><p style=\"margin-top:20px;font-size:14px;color:#6b7280;\">For security reasons, please change your password after your first login.</p></div>"}}'::jsonb,
  '{"required":["platformName","accountName","email","password"]}'::jsonb,
  86400
)
ON CONFLICT (notification_type) DO UPDATE SET
  name             = EXCLUDED.name,
  description      = EXCLUDED.description,
  template_group   = EXCLUDED.template_group,
  priority         = EXCLUDED.priority,
  default_content  = EXCLUDED.default_content,
  variables_schema = EXCLUDED.variables_schema,
  default_ttl      = EXCLUDED.default_ttl,
  is_system_managed = true,
  updated_at       = CURRENT_TIMESTAMP;

----------------------------------------------------------------------
-- PART 3: Amharic translations for all templates missing them
----------------------------------------------------------------------

-- 3a. system_announcement
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የስርዓት ማስታወቂያ: {{message}}',
  '{"in_app":{"title":"የስርዓት ማስታወቂያ","body":"{{message}}","actionUrl":"/announcements/{{slug}}"},"email":{"subject":"የስርዓት ማስታወቂያ: {{message}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">የስርዓት ማስታወቂያ</h2><p style=\"font-size:15px;line-height:1.7;\">{{message}}</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'system_announcement'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3b. policy_update
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'አስፈላጊ: የፖሊሲ ማሻሻያ',
  '{"in_app":{"title":"የፖሊሲ ማሻሻያ","body":"{{summary}}","actionUrl":"/policies/{{slug}}"},"email":{"subject":"አስፈላጊ: የፖሊሲ ማሻሻያ","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#d97706;\">የፖሊሲ ማሻሻያ</h2><p style=\"font-size:15px;line-height:1.7;\">{{summary}}</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'policy_update'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3c. welcome_message
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'እንኳን ወደ {{platformName}} በደህና መጡ',
  '{"in_app":{"title":"እንኳን ወደ {{platformName}} በደህና መጡ","body":"ሰላም {{accountName}}, እንኳን ወደ {{platformName}} በደህና መጡ! መመሪያዎችን በማሰስ ይጀምሩ።","actionUrl":"/guides"},"email":{"subject":"እንኳን ወደ {{platformName}} በደህና መጡ","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#16a34a;\">እንኳን ወደ {{platformName}} በደህና መጡ</h2><p style=\"font-size:15px;line-height:1.7;\">ሰላም {{accountName}},</p><p style=\"font-size:15px;line-height:1.7;\">እርስዎን በ{{platformName}} ውስጥ በማስተናገዳችን ደስተኞች ነን። መመሪያዎችን በመመርመር ይጀምሩ።</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'welcome_message'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3d. guide_deadline (in_app + email + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የማለቂያ ጊዜ እየተቃረበ ነው: {{stepTitle}}',
  '{"in_app":{"title":"የማለቂያ ጊዜ እየተቃረበ ነው","body":"{{stepTitle}} በ{{guideName}} ውስጥ በ{{deadlineDate}} ይጠናቀቃል","actionUrl":"/guides/{{guideSlug}}"},"email":{"subject":"የማለቂያ ጊዜ እየተቃረበ ነው: {{stepTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">የማለቂያ ጊዜ እየተቃረበ ነው</h2><p style=\"font-size:15px;line-height:1.7;\">ደረጃ <strong>{{stepTitle}}</strong> በ<strong>{{guideName}}</strong> ውስጥ በ<strong>{{deadlineDate}}</strong> ይጠናቀቃል።</p></div>"},"push":{"title":"የማለቂያ ጊዜ እየተቃረበ ነው","body":"{{stepTitle}} በ{{guideName}} ውስጥ በ{{deadlineDate}} ይጠናቀቃል"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'guide_deadline'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3e. guide_update
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'መመሪያ ተሻሽሏል',
  '{"in_app":{"title":"መመሪያ ተሻሽሏል","body":"{{guideName}} ተሻሽሏል","actionUrl":"/guides/{{guideSlug}}"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'guide_update'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3f. ai_quota_limit (in_app + email + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የAI አጠቃቀም ማንቂያ: {{percentUsed}}% ተጠቅመዋል',
  '{"in_app":{"title":"የAI ኮታ {{status}}","body":"የAI ኮታዎ {{percentUsed}}% ተጠቅመዋል። {{message}}","actionUrl":"/ai/usage"},"email":{"subject":"የAI አጠቃቀም ማንቂያ: {{percentUsed}}% ተጠቅመዋል","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#7c3aed;\">የAI አጠቃቀም ማንቂያ</h2><p style=\"font-size:15px;line-height:1.7;\">ከAI ኮታዎ <strong>{{percentUsed}}%</strong> ተጠቅመዋል።</p><p style=\"font-size:15px;line-height:1.7;\">{{message}}</p></div>"},"push":{"title":"የAI ኮታ {{status}}","body":"የAI ኮታዎ {{percentUsed}}% ተጠቅመዋል። {{message}}"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'ai_quota_limit'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3g. ai_response_ready (in_app + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የAI ምላሽ ዝግጁ ነው',
  '{"in_app":{"title":"የAI ምላሽ ዝግጁ ነው","body":"የእርስዎ የAI ጥያቄ \"{{queryPreview}}\" ተጠናቋል","actionUrl":"/ai/conversations/{{conversationId}}"},"push":{"title":"የAI ምላሽ ዝግጁ ነው","body":"የእርስዎ የAI ጥያቄ \"{{queryPreview}}\" ተጠናቋል"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'ai_response_ready'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3h. account_alert (in_app + email + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የደህንነት ማንቂያ: {{alertTitle}}',
  '{"in_app":{"title":"የደህንነት ማንቂያ","body":"{{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"የደህንነት ማንቂያ: {{alertTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">የደህንነት ማንቂያ</h2><div style=\"background:#fef2f2;border-left:4px solid #dc2626;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:15px;\"><strong>{{alertTitle}}</strong></p><p style=\"margin:12px 0 0;font-size:15px;line-height:1.7;\">{{alertMessage}}</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;color:#4b5563;\">ይህን እርምጃ ካልወሰዱ ወዲያውኑ መለያዎን ያስጠብቁ።</p></div>"},"push":{"title":"የደህንነት ማንቂያ","body":"{{alertMessage}}"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'account_alert'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3i. account_alert_critical (in_app + email + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  '[URGENT] የደህንነት ማንቂያ: {{alertTitle}}',
  '{"in_app":{"title":"የደህንነት ማንቂያ","body":"{{alertTitle}}: {{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"[URGENT] የደህንነት ማንቂያ: {{alertTitle}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#b91c1c;\">URGENT የደህንነት ማንቂያ</h2><div style=\"background:#fef2f2;border-left:4px solid #b91c1c;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:15px;\"><strong>{{alertTitle}}</strong></p><p style=\"margin:12px 0 0;font-size:15px;line-height:1.7;\">{{alertMessage}}</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;color:#4b5563;\">ይህን እርምጃ ካልወሰዱ ወዲያውኑ መለያዎን ያስጠብቁ።</p></div>"},"push":{"title":"የደህንነት ማንቂያ","body":"{{alertTitle}}: {{alertMessage}}"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'account_alert_critical'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3j. account_verification (uses {{otpCode}})
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የ{{platformName}} መለያዎን ያረጋግጡ',
  '{"in_app":{"title":"ማረጋገጫ ያስፈልጋል","body":"የማረጋገጫ ኮድዎ {{otpCode}} ነው። እባክዎ ይህን ኮድ ያስገቡ።","actionUrl":"/account/verify"},"email":{"subject":"የ{{platformName}} መለያዎን ያረጋግጡ","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">መለያዎን ያረጋግጡ</h2><p style=\"font-size:15px;line-height:1.7;\">የማረጋገጫ ኮድዎ <strong>{{otpCode}}</strong> ነው። እባክዎ ኢሜይልዎን ለማረጋገጥ ይህን ኮድ ያስገቡ።</p><p style=\"margin-top:20px;font-size:13px;color:#6b7280;\">ይህ የማረጋገጫ ሊንክ በ{{expiryMinutes}} ደቂቃ ውስጥ ያበቃል።</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'account_verification'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3k. password_reset
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የ{{platformName}} የይለፍ ቃልዎን ዳግም ያስጀምሩ',
  '{"email":{"subject":"የ{{platformName}} የይለፍ ቃልዎን ዳግም ያስጀምሩ","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#dc2626;\">የይለፍ ቃል ዳግም ማስጀመሪያ</h2><p style=\"font-size:15px;line-height:1.7;\">የ{{platformName}} መለያዎ የይለፍ ቃል እንዲዳግም ጥያቄ ደርሶናል።</p><p style=\"margin-top:20px;font-size:13px;color:#6b7280;\">ይህ ሊንክ በ{{expiryMinutes}} ደቂቃ ውስጥ ያበቃል።</p><p style=\"font-size:13px;color:#6b7280;\">ይህን ካልጠየቁ ይህን ኢሜይል ችላ ማለት ይችላሉ።</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'password_reset'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3l. payment_confirmation (in_app + email + push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'ክፍያ ተረጋግጧል: {{amount}} {{currency}}',
  '{"in_app":{"title":"ክፍያ ተረጋግጧል","body":"የ{{amount}} {{currency}} ክፍያዎ ተረጋግጧል።","actionUrl":"/account/billing"},"email":{"subject":"ክፍያ ተረጋግጧል: {{amount}} {{currency}}","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#16a34a;\">ክፍያ ተረጋግጧል</h2><div style=\"background:#f0fdf4;border-left:4px solid #16a34a;padding:16px;border-radius:8px;\"><p style=\"margin:0;font-size:16px;\"><strong>{{amount}} {{currency}}</strong></p><p style=\"margin:8px 0 0;font-size:14px;color:#374151;\">ክፍያዎ በተሳካ ሁኔታ ተከናውኗል።</p></div><p style=\"margin-top:20px;font-size:14px;line-height:1.7;\"><strong>ማጣቀሻ:</strong> {{referenceId}}<br/><strong>ቀን:</strong> {{paymentDate}}</p></div>"},"push":{"title":"ክፍያ ተረጋግጧል","body":"የ{{amount}} {{currency}} ክፍያዎ ተረጋግጧል።"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'payment_confirmation'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 3m. admin_provisioned (email only)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የ{{platformName}} አስተዳዳሪ መለያ አቅርቦት',
  '{"email":{"subject":"የ{{platformName}} አስተዳዳሪ መለያ አቅርቦት","body":"<div style=\"padding:24px;font-family:Arial,sans-serif;color:#111827;\"><h2 style=\"margin:0 0 16px;color:#2563eb;\">አስተዳዳሪ መለያ ተፈጥሯል</h2><p style=\"font-size:15px;line-height:1.7;\">ሰላም {{accountName}},</p><p style=\"font-size:15px;line-height:1.7;\">በ<strong>{{platformName}}</strong> ላይ የአስተዳዳሪ መለያ ተፈጥሯል።</p><div style=\"background:#f0f5ff;border-left:4px solid #2563eb;padding:16px;border-radius:8px;margin-top:16px;\"><p style=\"margin:0;font-size:15px;\"><strong>ኢሜይል:</strong> {{email}}</p><p style=\"margin:12px 0 0;font-size:15px;\"><strong>የይለፍ ቃል:</strong> {{password}}</p></div><p style=\"margin-top:20px;font-size:14px;color:#6b7280;\">እባክዎ ከመጀመሪያ ግቤትዎ በኋላ የይለፍ ቃልዎን ይቀይሩ።</p></div>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'admin_provisioned'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;
