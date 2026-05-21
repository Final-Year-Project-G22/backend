-- Seed compliance_info notification template with en + am translations
-- Apply after migration add_scheduled_alerts_compliance_and_guide_compliance_type

INSERT INTO notification_templates (id, name, description, notification_type, priority, is_system_managed, default_content, variables_schema)
VALUES (
  gen_random_uuid(),
  'Compliance Alert',
  'Notification when a compliance entry is about to expire.',
  'compliance_info', 1, true,
  '{"in_app":{"title":"{{complianceType}} Expiring","body":"Your {{complianceType}} expires in {{daysRemaining}} days. Renew before {{expiryDate}}."},"email":{"subject":"{{complianceType}} Expiring","body":"<p>Your <strong>{{complianceType}}</strong> expires in {{daysRemaining}} days.</p><p>Expiry date: {{expiryDate}}</p><p>Please renew before the deadline.</p>"},"push":{"title":"{{complianceType}} Expiring","body":"Your {{complianceType}} expires in {{daysRemaining}} days."}}'::jsonb,
  '{"required":["complianceType","daysRemaining","expiryDate"]}'::jsonb
)
ON CONFLICT (notification_type) DO UPDATE SET
  name             = EXCLUDED.name,
  description      = EXCLUDED.description,
  default_content  = EXCLUDED.default_content,
  variables_schema = EXCLUDED.variables_schema,
  updated_at       = CURRENT_TIMESTAMP;

-- Amharic translation
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'የማክበር ጊዜ ማስታወሻ',
  '{"in_app":{"title":"{{complianceType}} ጊዜው ያበቃል","body":"የ{{complianceType}} የማለቂያ ጊዜ በ{{daysRemaining}} ቀናት ውስጥ ያበቃል። እባክዎ ያድሱ።"},"email":{"subject":"{{complianceType}} የማለቂያ ጊዜ","body":"<p>የ<strong>{{complianceType}}</strong> የማለቂያ ጊዜ በ{{daysRemaining}} ቀናት ውስጥ ያበቃል።</p><p>የማለቂያ ቀን: {{expiryDate}}</p><p>እባክዎ ከመቆጠቡ በፊት ያድሱ።</p>"},"push":{"title":"{{complianceType}} ጊዜው ያበቃል","body":"የ{{complianceType}} የማለቂያ ጊዜ በ{{daysRemaining}} ቀናት ውስጥ ያበቃል።"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'compliance_info'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;
