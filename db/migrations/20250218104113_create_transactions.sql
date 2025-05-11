-- migrate:up
CREATE TABLE transactions (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    transaction_type VARCHAR(32) NOT NULL,
    description TEXT,
    spent_at TIMESTAMPTZ,
    amount INTEGER DEFAULT 0,
    is_shared BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT transactions_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT transactions_wallet_id_fk FOREIGN KEY (wallet_id) REFERENCES wallets(id)
);

CREATE TABLE transaction_shares (
    id VARCHAR(255) PRIMARY KEY,
    transaction_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    amount INTEGER DEFAULT 0,
    percentage INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT transaction_shares_transaction_id_fk FOREIGN KEY (transaction_id) REFERENCES transactions(id),
    CONSTRAINT transaction_shares_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);


-- migrate:down
DROP TABLE IF EXISTS transaction_shares;
DROP TABLE IF EXISTS transactions;
