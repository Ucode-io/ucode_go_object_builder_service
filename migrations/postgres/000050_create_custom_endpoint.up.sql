CREATE TABLE IF NOT EXISTS custom_endpoint (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT        NOT NULL,
    description    TEXT,
    sql_query      TEXT        NOT NULL,
    method         TEXT        NOT NULL DEFAULT 'POST',
    in_transaction BOOLEAN     NOT NULL DEFAULT false,
    custom_endpoint ADD COLUMN parameters JSONB DEFAULT '[]'
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);
