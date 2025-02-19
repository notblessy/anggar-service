-- migrate:up
CREATE TABLE budget_categories (
    id BIGSERIAL PRIMARY KEY,
    budget_id BIGINT,
    category VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT budget_categories_budget_id_fk FOREIGN KEY (budget_id) REFERENCES budgets(id)
);

-- migrate:down
DROP TABLE IF EXISTS budget_categories;
