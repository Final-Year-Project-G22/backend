# Payment Module — Database Migration & Seed

## Migration Files

Three new migration files will be added to the `migrations/` directory:

### Migration 1: `migrations/XXX_create_plans_table.up.sql`

```sql
CREATE TYPE plan_name AS ENUM ('Basic', 'Pro');
CREATE TYPE plan_period AS ENUM ('monthly', 'yearly');

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name plan_name NOT NULL,
    period plan_period NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    UNIQUE (name, period, deleted_at) WHERE deleted_at IS NULL
);

CREATE INDEX idx_plans_active ON plans (name, period) WHERE deleted_at IS NULL AND is_active = TRUE;
```

### Migration 2: `migrations/XXX_create_payments_table.up.sql`

```sql
CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed');

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    tx_ref VARCHAR(255) NOT NULL UNIQUE,
    chapa_ref VARCHAR(255),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    plan_name VARCHAR(50) NOT NULL,
    plan_period VARCHAR(20) NOT NULL,
    status payment_status NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50),
    checked_out_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_payments_tx_ref ON payments (tx_ref);
CREATE INDEX idx_payments_account_id ON payments (account_id);
CREATE INDEX idx_payments_status ON payments (status);
CREATE INDEX idx_payments_chapa_ref ON payments (chapa_ref);
CREATE INDEX idx_payments_account_status ON payments (account_id, status);
```

### Migration 3: `migrations/XXX_create_subscriptions_table.up.sql`

```sql
CREATE TYPE subscription_status AS ENUM ('active', 'expired', 'cancelled');

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    plan_name VARCHAR(50) NOT NULL,
    plan_period VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ETB',
    status subscription_status NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ NOT NULL,
    current_period_end TIMESTAMPTZ NOT NULL,
    cancelled_at TIMESTAMPTZ,
    renewal_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT chk_period_end_after_start CHECK (current_period_end > current_period_start),
    CONSTRAINT chk_cancelled_has_end CHECK (
        (status = 'cancelled' AND cancelled_at IS NOT NULL) OR
        (status != 'cancelled')
    )
);

-- Only one active subscription per account
CREATE UNIQUE INDEX idx_subscriptions_active
ON subscriptions (account_id)
WHERE status = 'active' AND deleted_at IS NULL;

CREATE INDEX idx_subscriptions_status ON subscriptions (status);
CREATE INDEX idx_subscriptions_account ON subscriptions (account_id);
CREATE INDEX idx_subscriptions_account_period
ON subscriptions (account_id, current_period_end);
```

> **Note:** The `account_id` FK references `accounts(id)`. The actual accounts table name and structure depend on the IAM module — verify during implementation. If the IAM module uses a different table name (e.g., `users`), adjust the FK accordingly. A view or wrapper may be needed.

### Migration 3b: Add foreign key from payments to subscriptions

This requires subscription_id to exist after Migration 3. This may need to be a separate migration run after 3, or the FK may be added as a deferred constraint.

---

## Seed Data

A seed migration or seed script will insert the 4 static plans:

```sql
INSERT INTO plans (name, period, amount, currency, is_active) VALUES
    ('Basic',  'monthly', 9900,  'ETB', TRUE),
    ('Basic',  'yearly',  95000, 'ETB', TRUE),
    ('Pro',    'monthly', 19900, 'ETB', TRUE),
    ('Pro',    'yearly',  190000,'ETB', TRUE)
ON CONFLICT (name, period) DO UPDATE SET
    amount = EXCLUDED.amount,
    is_active = EXCLUDED.is_active;
```

---

## Schema Registration

Following the project's Atlas migration pattern:

1. Create GORM entity structs in `internal/modules/payment/domain/entity/`
2. Register them via `EntityProvider` in `internal/modules/payment/entities.go`
3. The `SchemaManager` will detect new entities and generate HCL schema diffs
4. Run `atlas migrate diff` to generate the migration SQL files above

---

## Rollback Migrations

Each `.up.sql` file needs a corresponding `.down.sql` file:

- `down.sql` for plans: `DROP TABLE IF EXISTS plans;` (and `DROP TYPE IF EXISTS plan_name, plan_period`)
- `down.sql` for payments: `DROP TABLE IF EXISTS payments;` (and `DROP TYPE IF EXISTS payment_status`)
- `down.sql` for subscriptions: `DROP TABLE IF EXISTS subscriptions;` (and `DROP TYPE IF EXISTS subscription_status`)
- `down.sql` for FK: Remove the FK constraint from payments