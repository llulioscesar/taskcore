CREATE TABLE filters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    jql TEXT NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    is_favorite BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, name)
);

CREATE INDEX idx_filters_owner ON filters(owner_id);
CREATE INDEX idx_filters_public ON filters(is_public) WHERE is_public = TRUE;

CREATE TABLE filter_shares (
    filter_id UUID NOT NULL REFERENCES filters(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES groups(id) ON DELETE CASCADE,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id IS NOT NULL OR group_id IS NOT NULL OR project_id IS NOT NULL)
);

CREATE INDEX idx_filter_shares_filter ON filter_shares(filter_id);
CREATE INDEX idx_filter_shares_user ON filter_shares(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_filter_shares_group ON filter_shares(group_id) WHERE group_id IS NOT NULL;

CREATE TABLE filter_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filter_id UUID NOT NULL REFERENCES filters(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    schedule VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (filter_id, user_id)
);

CREATE INDEX idx_filter_subscriptions_user ON filter_subscriptions(user_id);
