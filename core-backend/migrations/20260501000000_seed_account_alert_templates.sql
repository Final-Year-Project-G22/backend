-- Seed critical and informational account alert notification templates
-- Uses ON CONFLICT to be idempotent

INSERT INTO notification_templates (id, name, description, notification_type, category, priority, is_system_managed, default_content, variables_schema, default_ttl)
VALUES
(
  gen_random_uuid(),
  'Critical Account Alert',
  'Mandatory account security alert that bypasses user opt-out and quiet hours.',
  'account_alert_critical', 'security', 2, true,
  '{"inapp":{"title":"Security Alert","body":"{{alertTitle}}: {{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"[URGENT] Security Alert: {{alertTitle}}","body":"<p><strong>{{alertTitle}}</strong></p><p>{{alertMessage}}</p><p>If you did not perform this action, secure your account immediately: <a href=\"{{securityUrl}}\">Security Settings</a></p>"}}'::jsonb,
  '{"required":["alertTitle","alertMessage","alertCode","securityUrl"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Informational Account Alert',
  'Informational account security notification that respects user preferences and quiet hours.',
  'account_alert_info', 'security', 1, true,
  '{"inapp":{"title":"Account Notice","body":"{{alertTitle}}: {{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"Account Notice: {{alertTitle}}","body":"<p><strong>{{alertTitle}}</strong></p><p>{{alertMessage}}</p><p><a href=\"{{securityUrl}}\">Security Settings</a></p>"}}'::jsonb,
  '{"required":["alertTitle","alertMessage","alertCode","securityUrl"]}'::jsonb,
  NULL
)
ON CONFLICT (notification_type) DO UPDATE SET
  name             = EXCLUDED.name,
  description      = EXCLUDED.description,
  category         = EXCLUDED.category,
  priority         = EXCLUDED.priority,
  default_content  = EXCLUDED.default_content,
  variables_schema = EXCLUDED.variables_schema,
  default_ttl      = EXCLUDED.default_ttl,
  is_system_managed = true,
  updated_at       = CURRENT_TIMESTAMP;
