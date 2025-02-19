-- migrate:up
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    wallet_id BIGINT,
    budget_id BIGINT,
    category VARCHAR(255) NOT NULL,
    transaction_type VARCHAR(32) NOT NULL,
    description TEXT,
    spent_at TIMESTAMPTZ,
    amount INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT transactions_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT transactions_wallet_id_fk FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    CONSTRAINT transactions_budget_id_fk FOREIGN KEY (budget_id) REFERENCES budgets(id)
);


-- migrate:down
DROP TABLE IF EXISTS transactions;
