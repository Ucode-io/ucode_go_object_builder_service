CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS "app" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    tables JSONB,
    icon TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "table" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    label TEXT,
    description TEXT,
    subtitle_field_slug VARCHAR(255),
    is_changed BOOLEAN DEFAULT false,
    icon TEXT,
    show_in_menu BOOLEAN DEFAULT true,
    commit_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);



CREATE TYPE "relation_type" AS ENUM (
    'One2One',
    'One2Many',
    'Many2One',
    'Many2Many',
    'Recursive',
    'Many2Dynamic'
);

CREATE TABLE IF NOT EXISTS "relation" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_from VARCHAR(255) NOT NULL,
    table_to VARCHAR(255),
    field_from VARCHAR(255) NOT NULL,
    field_to VARCHAR(255) NOT NULL,
    "type" relation_type NOT NULL,
    view_fields TEXT [],
    relation_field_slug VARCHAR(255),
    dynamic_tables JSONB,
    editable BOOLEAN,
    auto_filters JSONB,
    is_user_id_default BOOLEAN,
    cascadings JSONB,
    object_id_from_jwt BOOLEAN,
    cascading_tree_table_slug VARCHAR(512),
    cascading_tree_field_slug VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
CREATE TABLE IF NOT EXISTS "field" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_id UUID REFERENCES "table"("id") ON DELETE CASCADE,
    "required" BOOLEAN DEFAULT false,
    slug VARCHAR(255) NOT NULL,
    label TEXT NOT NULL,
    "default" VARCHAR(255),
    "type" VARCHAR(255),
    "index" VARCHAR(255),
    attributes JSONB,
    is_visible BOOLEAN DEFAULT true,
    autofill_field VARCHAR(512),
    autofill_table VARCHAR(512),
    relation_id VARCHAR(255),
    commit_id VARCHAR(255),
    "unique" BOOLEAN,
    "automatic" BOOLEAN,
    icon TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "dashboard" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    icon TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "panel" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    query TEXT NOT NULL,
    coordinates FLOAT [] NOT NULL,
    attributes JSONB,
    dashboard_id UUID REFERENCES "dashboard"("id"),
    has_pagination BOOLEAN NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TYPE variable_type AS ENUM ('QUERY', 'CUSTOM');

CREATE TABLE IF NOT EXISTS "variables" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(255) NOT NULL,
    "type" variable_type NOT NULL,
    label VARCHAR(512) NOT NULL,
    dashboard_id UUID REFERENCES "dashboard"("id"),
    field_slug VARCHAR(255),
    options TEXT [],
    view_field_slug VARCHAR(255),
    query TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "custom_event" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_slug VARCHAR(255) NOT NULL,
    "type" variable_type NOT NULL,
    icon TEXT,
    label TEXT NOT NULL,
    event_path VARCHAR(255),
    "url" TEXT,
    "disable" BOOLEAN DEFAULT false,
    method VARCHAR(255),
    action_type VARCHAR(255),
    attributes JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "project" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain TEXT,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "client_platform" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID,
    name VARCHAR(255),
    subdomain TEXT,
    client_type_ids TEXT [],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "client_type" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID,
    name VARCHAR(255),
    self_register BOOLEAN DEFAULT false,
    self_recover BOOLEAN DEFAULT false,
    client_platform_ids TEXT [],
    confirm_by TEXT [],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "role" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID,
    client_platform_id UUID,
    client_type_id UUID,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "user" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID,
    client_platform_id UUID,
    client_type_id UUID,
    role_id UUID,
    active FLOAT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "test_login" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    login_strategy VARCHAR(255),
    table_slug VARCHAR(255),
    login_label VARCHAR(255),
    login_view VARCHAR(255),
    password_view VARCHAR(255),
    password_label VARCHAR(255),
    object_id VARCHAR(255),
    client_type_id UUID REFERENCES "client_type"("guid") ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "connections" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_slug VARCHAR(255),
    view_slug VARCHAR(255),
    view_label VARCHAR(255),
    name VARCHAR(255),
    type JSONB,
    icon TEXT,
    main_table_slug VARCHAR(255),
    field_slug VARCHAR(255),
    client_type_id UUID REFERENCES "client_type"("guid") ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "record_permission" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES "role"("guid") ON DELETE CASCADE,
    "read" VARCHAR(255),
    "write" VARCHAR(255),
    "update" VARCHAR(255),
    "delete" VARCHAR(255),
    "is_public" BOOLEAN DEFAULT false,
    "is_have_condition" BOOLEAN DEFAULT false,
    table_slug VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "field_permission" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES "role"("guid") ON DELETE CASCADE,
    "label" VARCHAR(255),
    "table_slug" VARCHAR(255),
    "field_id" UUID REFERENCES "field"("id") ON DELETE CASCADE,
    "edit_permission" BOOLEAN DEFAULT true,
    "view_permission" BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "action_permission" (
    guid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES "role"("guid") ON DELETE CASCADE,
    "table_slug" VARCHAR(255),
    "custom_event_id" UUID REFERENCES "custom_event"("id") ON DELETE CASCADE,
    "permission" BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
