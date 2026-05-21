-- Seed scheduled alert templates for user-created notifications
-- Apply after migration add_scheduled_alerts_compliance_and_guide_compliance_type

INSERT INTO scheduled_alert_templates (id, slug, name, default_title, default_body, default_channel, is_active)
VALUES
  (gen_random_uuid(), 'custom',               'Custom',              '',                                       '',                                                                      NULL,     true),
  (gen_random_uuid(), 'tax_filing',           'Tax Filing Reminder', 'Tax Filing Due',                        'Your tax filing deadline is approaching. Make sure to submit your returns on time.',               'in_app', true),
  (gen_random_uuid(), 'license_renewal',      'License Renewal',     'License Expiring',                      'Your trade license renewal is due soon. Please prepare the required documents.',                    'in_app', true),
  (gen_random_uuid(), 'registration_renewal', 'Registration Renewal', 'Registration Expiring',                'Your business registration renewal is approaching. Check the expiry date and renew on time.',       'in_app', true),
  (gen_random_uuid(), 'meeting',              'Meeting Reminder',    'Meeting Today',                         'Don''t forget your scheduled meeting.',                                                             'push',   true),
  (gen_random_uuid(), 'deadline',             'Custom Deadline',     'Deadline Approaching',                  'A deadline you set is coming up.',                                                                  'in_app', true)
ON CONFLICT (slug) DO UPDATE SET
  name             = EXCLUDED.name,
  default_title    = EXCLUDED.default_title,
  default_body     = EXCLUDED.default_body,
  default_channel  = EXCLUDED.default_channel,
  is_active        = EXCLUDED.is_active,
  updated_at       = CURRENT_TIMESTAMP;
