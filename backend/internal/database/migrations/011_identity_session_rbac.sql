-- M04 identity and authorization storage foundation.
-- This migration creates schema only: no auth routes, accounts or sessions are exposed.
CREATE TABLE roles (
    role_key TEXT PRIMARY KEY CHECK (role_key IN ('customer', 'seller', 'admin')),
    display_name TEXT NOT NULL CHECK (length(btrim(display_name)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO roles (role_key, display_name) VALUES
    ('customer', 'Customer'), ('seller', 'Seller'), ('admin', 'Administrator');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_normalized TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending_verification'
        CHECK (status IN ('pending_verification', 'active', 'disabled')),
    email_verified_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_normalized_check CHECK (
        email_normalized = lower(btrim(email_normalized))
        AND email_normalized !~ '[[:space:]]'
        AND position('@' IN email_normalized) > 1
        AND position('@' IN email_normalized) < length(email_normalized)
    ),
    CONSTRAINT users_password_argon2id_check CHECK (
        password_hash LIKE '$argon2id$%' AND length(password_hash) BETWEEN 32 AND 512
    ),
    CONSTRAINT users_status_lifecycle_check CHECK (
        (status = 'disabled') = (disabled_at IS NOT NULL)
        AND (status <> 'active' OR email_verified_at IS NOT NULL)
    )
);
CREATE UNIQUE INDEX users_email_normalized_uidx ON users(email_normalized);
CREATE INDEX users_status_created_idx ON users(status, created_at DESC);

CREATE TABLE user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    role_key TEXT NOT NULL REFERENCES roles(role_key) ON DELETE RESTRICT,
    granted_by_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at TIMESTAMPTZ,
    CONSTRAINT user_roles_not_self_granted_check
        CHECK (granted_by_user_id IS NULL OR granted_by_user_id <> user_id),
    CONSTRAINT user_roles_revocation_check
        CHECK (revoked_at IS NULL OR revoked_at >= granted_at)
);
CREATE UNIQUE INDEX user_roles_one_active_uidx
    ON user_roles(user_id, role_key) WHERE revoked_at IS NULL;
CREATE INDEX user_roles_user_active_idx
    ON user_roles(user_id, role_key) WHERE revoked_at IS NULL;

-- Every new account receives customer only. Privileged roles are separate grants.
CREATE OR REPLACE FUNCTION assign_default_customer_role()
RETURNS trigger
LANGUAGE plpgsql
AS $customer_role$
BEGIN
    INSERT INTO user_roles(user_id, role_key) VALUES (NEW.id, 'customer');
    RETURN NEW;
END;
$customer_role$;
CREATE TRIGGER users_default_customer_role
    AFTER INSERT ON users
    FOR EACH ROW EXECUTE FUNCTION assign_default_customer_role();

CREATE TABLE seller_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    vendor_id BIGINT NOT NULL REFERENCES vendors(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended', 'revoked')),
    reviewed_by_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    activated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT seller_memberships_activation_check
        CHECK (status <> 'active' OR activated_at IS NOT NULL),
    UNIQUE(user_id, vendor_id)
);
CREATE INDEX seller_memberships_active_user_idx
    ON seller_memberships(user_id, vendor_id) WHERE status = 'active';
CREATE INDEX seller_memberships_vendor_status_idx
    ON seller_memberships(vendor_id, status, created_at DESC);

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    token_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    csrf_token_digest BYTEA NOT NULL CHECK (octet_length(csrf_token_digest) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    idle_expires_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT CHECK (revocation_reason IS NULL OR revocation_reason IN (
        'logout', 'password_changed', 'account_disabled',
        'privilege_changed', 'admin_revocation', 'security_rotation'
    )),
    CONSTRAINT user_sessions_expiry_check CHECK (
        last_seen_at >= created_at AND idle_expires_at > last_seen_at
        AND idle_expires_at <= expires_at AND expires_at > created_at
    ),
    CONSTRAINT user_sessions_revocation_check CHECK (
        (revoked_at IS NULL) = (revocation_reason IS NULL)
        AND (revoked_at IS NULL OR revoked_at >= created_at)
    )
);
CREATE INDEX user_sessions_user_active_idx
    ON user_sessions(user_id, expires_at DESC) WHERE revoked_at IS NULL;
CREATE INDEX user_sessions_expiry_idx
    ON user_sessions(expires_at) WHERE revoked_at IS NULL;

CREATE TABLE auth_action_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    purpose TEXT NOT NULL CHECK (purpose IN ('email_verification', 'password_reset')),
    token_digest BYTEA NOT NULL UNIQUE CHECK (octet_length(token_digest) = 32),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    CONSTRAINT auth_action_tokens_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT auth_action_tokens_single_use_check CHECK (
        (consumed_at IS NULL OR consumed_at >= created_at)
        AND (revoked_at IS NULL OR revoked_at >= created_at)
        AND NOT (consumed_at IS NOT NULL AND revoked_at IS NOT NULL)
    )
);
CREATE UNIQUE INDEX auth_action_tokens_one_active_purpose_uidx
    ON auth_action_tokens(user_id, purpose)
    WHERE consumed_at IS NULL AND revoked_at IS NULL;
CREATE INDEX auth_action_tokens_expiry_idx
    ON auth_action_tokens(expires_at) WHERE consumed_at IS NULL AND revoked_at IS NULL;

CREATE TABLE audit_events (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL CHECK (event_type ~ '^[a-z][a-z0-9_.-]{1,95}$'),
    actor_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    subject_user_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    resource_type TEXT CHECK (resource_type IS NULL OR length(resource_type) BETWEEN 1 AND 80),
    resource_id TEXT CHECK (resource_id IS NULL OR length(resource_id) BETWEEN 1 AND 160),
    request_id TEXT CHECK (request_id IS NULL OR length(request_id) BETWEEN 1 AND 128),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT audit_events_metadata_check CHECK (
        jsonb_typeof(metadata) = 'object' AND pg_column_size(metadata) <= 4096
        AND NOT (metadata ?| ARRAY[
            'password', 'password_hash', 'session_token', 'token',
            'token_digest', 'csrf_token', 'authorization', 'cookie',
            'secret', 'api_key', 'card_number'
        ])
    )
);
CREATE INDEX audit_events_type_created_idx ON audit_events(event_type, created_at DESC, id DESC);
CREATE INDEX audit_events_actor_created_idx ON audit_events(actor_user_id, created_at DESC, id DESC);
CREATE INDEX audit_events_subject_created_idx ON audit_events(subject_user_id, created_at DESC, id DESC);

-- Enforce append-only audit history at the database boundary.
CREATE OR REPLACE FUNCTION reject_audit_event_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $audit_guard$
BEGIN
    RAISE EXCEPTION 'audit_events is append-only' USING ERRCODE = '55000';
END;
$audit_guard$;
CREATE TRIGGER audit_events_no_update_delete
    BEFORE UPDATE OR DELETE ON audit_events
    FOR EACH ROW EXECUTE FUNCTION reject_audit_event_mutation();
CREATE TRIGGER audit_events_no_truncate
    BEFORE TRUNCATE ON audit_events
    FOR EACH STATEMENT EXECUTE FUNCTION reject_audit_event_mutation();
