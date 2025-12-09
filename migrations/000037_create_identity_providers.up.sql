-- Identity Providers (OpenID Connect configuration)
CREATE TABLE identity_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) NOT NULL UNIQUE,
    issuer VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret VARCHAR(500) NOT NULL,
    scopes TEXT[] NOT NULL DEFAULT '{openid,email,profile}',
    auto_create_users BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_identity_providers_slug ON identity_providers(slug);
CREATE INDEX idx_identity_providers_active ON identity_providers(is_active) WHERE is_active = true;

-- User Identities (link between users and identity providers)
CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES identity_providers(id) ON DELETE CASCADE,
    subject VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    picture VARCHAR(500),
    raw_claims JSONB,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider_id, subject)
);

CREATE INDEX idx_user_identities_user ON user_identities(user_id);
CREATE INDEX idx_user_identities_provider ON user_identities(provider_id);
CREATE INDEX idx_user_identities_subject ON user_identities(provider_id, subject);

-- OIDC State (for OAuth flow CSRF protection)
CREATE TABLE oidc_states (
    state VARCHAR(64) PRIMARY KEY,
    provider_id UUID NOT NULL REFERENCES identity_providers(id) ON DELETE CASCADE,
    redirect_uri VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '10 minutes'
);

CREATE INDEX idx_oidc_states_expires ON oidc_states(expires_at);
