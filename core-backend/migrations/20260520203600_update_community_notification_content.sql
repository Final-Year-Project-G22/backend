-- Update community_reply notification template: add push content, remove email
-- Replies deliver via in-app + push only (no email)
-- Seed Amharic translations for all community notification templates

-- 1. Update community_reply default content: add push, remove email
UPDATE notification_templates
SET default_content = '{"inapp":{"title":"New Reply","body":"{{authorName}} replied to {{threadTitle}}","actionUrl":"/community/threads/{{threadSlug}}"},"push":{"title":"New Reply","body":"{{authorName}} replied to {{threadTitle}}"}}'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'community_reply';

-- 2. Update community_solution default content: add push
UPDATE notification_templates
SET default_content = '{"inapp":{"title":"Solution Found","body":"A solution was marked on {{threadTitle}}","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"Solution posted for {{threadTitle}}","body":"<p>A solution was posted for your thread <strong>{{threadTitle}}</strong>.</p><p><a href=\"{{threadUrl}}\">View solution</a></p>"},"push":{"title":"Solution Found","body":"A solution was marked on {{threadTitle}}"}}'::jsonb,
    updated_at = CURRENT_TIMESTAMP
WHERE notification_type = 'community_solution';

-- 3. Amharic translation for community_reply (no email, push + in-app)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  '{{authorName}} በ{{threadTitle}} ላይ ምላሽ ሰጥተዋል',
  '{"in_app":{"title":"አዲስ ምላሽ","body":"{{authorName}} በ{{threadTitle}} ላይ ምላሽ ሰጥተዋል","actionUrl":"/community/threads/{{threadSlug}}"},"push":{"title":"አዲስ ምላሽ","body":"{{authorName}} በ{{threadTitle}} ላይ ምላሽ ሰጥተዋል"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'community_reply'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 4. Amharic translation for community_solution (add push)
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  'ለ{{threadTitle}} መፍትሄ ተለጠፈ',
  '{"in_app":{"title":"መፍትሄ ተገኝቷል","body":"በ{{threadTitle}} ላይ መፍትሄ ተምልክቷል","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"ለ{{threadTitle}} መፍትሄ ተለጠፈ","body":"<p>ለእርስዎ ርዕስ <strong>{{threadTitle}}</strong> መፍትሄ ተለጠፈ።</p><p><a href=\"{{threadUrl}}\">መፍትሄውን ይመልከቱ</a></p>"},"push":{"title":"መፍትሄ ተገኝቷል","body":"በ{{threadTitle}} ላይ መፍትሄ ተምልክቷል"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'community_solution'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;

-- 5. Amharic translation for community_mention
INSERT INTO notification_template_translations (id, template_id, language, subject, content)
SELECT
  gen_random_uuid(),
  t.id,
  'am',
  '{{authorName}} በ{{threadTitle}} ውስጥ ጠቅሰውዎታል',
  '{"in_app":{"title":"ተጠቅሰዋል","body":"{{authorName}} በ{{threadTitle}} ውስጥ ጠቅሰውዎታል","actionUrl":"/community/threads/{{threadSlug}}"},"email":{"subject":"{{authorName}} በ{{threadTitle}} ውስጥ ጠቅሰውዎታል","body":"<p>{{authorName}} በ<strong>{{threadTitle}}</strong> ውስጥ ጠቅሰውዎታል፦</p><blockquote>{{mentionExcerpt}}</blockquote><p><a href=\"{{threadUrl}}\">ልጥፉን ይመልከቱ</a></p>"}}'::jsonb
FROM notification_templates t
WHERE t.notification_type = 'community_mention'
ON CONFLICT (template_id, language) DO UPDATE SET
  subject    = EXCLUDED.subject,
  content    = EXCLUDED.content,
  updated_at = CURRENT_TIMESTAMP;
