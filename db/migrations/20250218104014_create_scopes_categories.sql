-- migrate:up
CREATE TABLE scope_categories (
    id BIGSERIAL PRIMARY KEY,
    scope_id BIGINT,
    category VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT scope_categories_scope_id_fk FOREIGN KEY (scope_id) REFERENCES scopes(id)
);

-- migrate:down
DROP TABLE IF EXISTS scope_categories;
