-- migrate:up
CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255),
    name VARCHAR(255) NOT NULL,
    balance INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT wallets_user_id_fk FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT wallets_name_idx UNIQUE (name)
);

-- migrate:down
DROP TABLE IF EXISTS wallets;
