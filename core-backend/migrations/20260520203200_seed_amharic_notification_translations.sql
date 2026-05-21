-- Seed Amharic translations for guide_step_completed and account_alert_info templates
-- Apply after migration add_scheduled_alerts_compliance_and_guide_compliance_type

-- guide_step_completed → Amharic
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'ደረጃ ተጠናቋል',
  '{"in_app":{"title":"ደረጃ ተጠናቋል","body":"\"{{stepTitle}}\" ደረጃን በ{{guideName}} ውስጥ አጠናቀዋል","actionUrl":"/guides/{{guideSlug}}"},"email":{"subject":"ደረጃ ተጠናቋል: {{stepTitle}}","body":"<p>\"{{stepTitle}}\" ደረጃን በ<strong>{{guideName}}</strong> ውስጥ አጠናቀዋል።</p>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'guide_step_completed'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- account_alert_info → Amharic
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የመከበር ጊዜ ማስታወሻ',
  '{"in_app":{"title":"የ{{alertTitle}} ማስታወሻ","body":"{{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"የመለያ ማስታወሻ: {{alertTitle}}","body":"<p><strong>{{alertTitle}}</strong></p><p>{{alertMessage}}</p><p><a href=\"{{securityUrl}}\">የደህንነት ቅንብሮች</a></p>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'account_alert_info'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;
