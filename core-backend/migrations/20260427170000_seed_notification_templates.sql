-- Seed 15 system-managed notification templates
-- Uses ON CONFLICT to be idempotent

INSERT INTO notification_templates (id, name, description, notification_type, category, priority, is_system_managed, default_content, variables_schema, default_ttl)
VALUES
(
  gen_random_uuid(),
  'System Announcement',
  'Platform-wide announcements from the system administrators.',
  'system_announcement', 'system', 1, true,
  '{"inapp":{"title":"System Announcement","body":"{{message}}","actionUrl":"/announcements/{{slug}}"},"email":{"subject":"System Announcement: {{message}}","body":"<p>{{message}}</p><p>View details: <a href=\"{{url}}\">{{url}}</a></p>"}}'::jsonb,
  '{"required":["message","slug","url"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Policy Update',
  'Notification when platform policies or terms of service are updated.',
  'policy_update', 'system', 1, true,
  '{"inapp":{"title":"Policy Update","body":"{{summary}}","actionUrl":"/policies/{{slug}}"},"email":{"subject":"Important: Policy Update","body":"<p>{{summary}}</p><p>Review the updated policy: <a href=\"{{url}}\">{{url}}</a></p>"}}'::jsonb,
  '{"required":["summary","slug","url"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Welcome Message',
  'Welcome notification sent to new accounts after registration.',
  'welcome_message', 'system', 1, true,
  '{"inapp":{"title":"Welcome to {{platformName}}","body":"Hi {{accountName}}, welcome! Get started by exploring our guides.","actionUrl":"/guides"},"email":{"subject":"Welcome to {{platformName}}","body":"<p>Hi {{accountName}},</p><p>Welcome to {{platformName}}! We''re excited to have you on board.</p><p>Get started: <a href=\"{{gettingStartedUrl}}\">Begin here</a></p>"}}'::jsonb,
  '{"required":["platformName","accountName","gettingStartedUrl"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Community Reply',
  'Notification when someone replies to your thread.',
  'community_reply', 'community', 0, true,
  '{"inapp":{"title":"New Reply","body":"{{authorName}} replied to {{threadTitle}}","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"{{authorName}} replied to {{threadTitle}}","body":"<p>{{authorName}} replied to your thread <strong>{{threadTitle}}</strong>:</p><blockquote>{{replyExcerpt}}</blockquote><p><a href=\"{{threadUrl}}\">View reply</a></p>"}}'::jsonb,
  '{"required":["authorName","threadTitle","threadSlug"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Community Solution',
  'Notification when your thread receives a solution.',
  'community_solution', 'community', 1, true,
  '{"inapp":{"title":"Solution Found","body":"A solution was marked on {{threadTitle}}","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"Solution posted for {{threadTitle}}","body":"<p>A solution was posted for your thread <strong>{{threadTitle}}</strong>.</p><p><a href=\"{{threadUrl}}\">View solution</a></p>"}}'::jsonb,
  '{"required":["threadTitle","threadSlug","threadUrl"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Community Mention',
  'Notification when you are mentioned in a post.',
  'community_mention', 'community', 0, true,
  '{"inapp":{"title":"You Were Mentioned","body":"{{authorName}} mentioned you in {{threadTitle}}","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"{{authorName}} mentioned you in {{threadTitle}}","body":"<p>{{authorName}} mentioned you in <strong>{{threadTitle}}</strong>:</p><blockquote>{{mentionExcerpt}}</blockquote><p><a href=\"{{threadUrl}}\">View post</a></p>"}}'::jsonb,
  '{"required":["authorName","threadTitle","threadSlug"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Guide Step Completed',
  'Notification when a guide step is completed.',
  'guide_step_completed', 'guide', 1, true,
  '{"inapp":{"title":"Step Completed","body":"You completed \"{{stepTitle}}\" in {{guideName}}","actionUrl":"/guides/{{guideSlug}}"}}'::jsonb,
  '{"required":["stepTitle","guideName","guideSlug"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Guide Deadline',
  'Notification when a compliance deadline is approaching.',
  'guide_deadline', 'guide', 1, true,
  '{"inapp":{"title":"Deadline Approaching","body":"{{stepTitle}} in {{guideName}} is due by {{deadlineDate}}","actionUrl":"/guides/{{guideSlug}}"},"email":{"subject":"Deadline Approaching: {{stepTitle}}","body":"<p>The step <strong>{{stepTitle}}</strong> in <strong>{{guideName}}</strong> is due by {{deadlineDate}}.</p><p><a href=\"{{guideUrl}}\">View step</a></p>"}}'::jsonb,
  '{"required":["stepTitle","guideName","guideSlug","deadlineDate"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Guide Update',
  'Notification when a guide you follow is updated.',
  'guide_update', 'guide', 0, true,
  '{"inapp":{"title":"Guide Updated","body":"{{guideName}} has been updated","actionUrl":"/guides/{{guideSlug}}"}}'::jsonb,
  '{"required":["guideName","guideSlug"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'AI Quota Limit',
  'Notification when AI usage quota is reached or near limit.',
  'ai_quota_limit', 'ai', 1, true,
  '{"inapp":{"title":"AI Quota {{status}}","body":"You have {{percentUsed}}% of your AI quota. {{message}}","actionUrl":"/ai/usage"},"email":{"subject":"AI Usage Alert: {{percentUsed}}% Used","body":"<p>You have used <strong>{{percentUsed}}%</strong> of your AI quota.</p><p>{{message}}</p><p><a href=\"{{usageUrl}}\">View usage</a></p>"}}'::jsonb,
  '{"required":["status","percentUsed","message"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'AI Response Ready',
  'Notification when an async AI response is ready.',
  'ai_response_ready', 'ai', 0, true,
  '{"inapp":{"title":"AI Response Ready","body":"Your AI request \"{{queryPreview}}\" has completed","actionUrl":"/ai/conversations/{{conversationId}}"}}'::jsonb,
  '{"required":["queryPreview","conversationId"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Account Alert',
  'Security alert notification for account activities.',
  'account_alert', 'security', 2, true,
  '{"inapp":{"title":"Security Alert","body":"{{alertMessage}}","actionUrl":"/account/security"},"email":{"subject":"Security Alert: {{alertTitle}}","body":"<p><strong>{{alertTitle}}</strong></p><p>{{alertMessage}}</p><p>If you did not perform this action, secure your account immediately: <a href=\"{{securityUrl}}\">Security Settings</a></p>"}}'::jsonb,
  '{"required":["alertTitle","alertMessage"]}'::jsonb,
  NULL
),
(
  gen_random_uuid(),
  'Account Verification',
  'Verification email for account operations.',
  'account_verification', 'security', 2, true,
  '{"inapp":{"title":"Verification Required","body":"{{verificationMessage}}","actionUrl":"/account/verify"},"email":{"subject":"Verify your {{platformName}} account","body":"<p>{{verificationMessage}}</p><p><a href=\"{{verificationUrl}}\">Verify Now</a></p><p>This link expires in {{expiryMinutes}} minutes.</p>"}}'::jsonb,
  '{"required":["platformName","verificationMessage","verificationUrl","expiryMinutes"]}'::jsonb,
  3600
),
(
  gen_random_uuid(),
  'Password Reset',
  'Password reset request notification.',
  'password_reset', 'security', 3, true,
  '{"email":{"subject":"Reset your {{platformName}} password","body":"<p>We received a password reset request for your {{platformName}} account.</p><p><a href=\"{{resetUrl}}\">Reset Password</a></p><p>This link expires in {{expiryMinutes}} minutes. If you did not request this, please ignore this email.</p>"}}'::jsonb,
  '{"required":["platformName","resetUrl","expiryMinutes"]}'::jsonb,
  3600
),
(
  gen_random_uuid(),
  'Payment Confirmation',
  'Notification confirming a payment transaction.',
  'payment_confirmation', 'payment', 1, true,
  '{"inapp":{"title":"Payment Confirmed","body":"Your payment of {{amount}} {{currency}} has been confirmed.","actionUrl":"/account/billing"},"email":{"subject":"Payment Confirmed: {{amount}} {{currency}}","body":"<p>Your payment of <strong>{{amount}} {{currency}}</strong> has been confirmed.</p><p>Reference: {{referenceId}}</p><p>Date: {{paymentDate}}</p><p><a href=\"{{billingUrl}}\">View billing details</a></p>"}}'::jsonb,
  '{"required":["amount","currency","referenceId","paymentDate"]}'::jsonb,
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
