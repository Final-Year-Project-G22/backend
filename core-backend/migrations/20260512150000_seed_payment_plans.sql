-- Seed subscription plans.
INSERT INTO plans (name, period, amount, currency, is_active) VALUES
  ('Basic', 'monthly', 0, 'ETB', true),
  ('Pro',   'monthly', 100, 'ETB', true),
  ('Pro',   'yearly',  1000, 'ETB', true)
ON CONFLICT (name, period) DO UPDATE SET
  amount    = EXCLUDED.amount,
  is_active = EXCLUDED.is_active;
