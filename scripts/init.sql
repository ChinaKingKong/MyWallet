CREATE TABLE IF NOT EXISTS wallets (
    id VARCHAR(64) PRIMARY KEY,
    address VARCHAR(64) UNIQUE NOT NULL,
    balance DECIMAL(20,8) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(64) PRIMARY KEY,
    from_wallet VARCHAR(64) NOT NULL,
    to_wallet VARCHAR(64) NOT NULL,
    amount DECIMAL(20,8) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    FOREIGN KEY (from_wallet) REFERENCES wallets(address),
    FOREIGN KEY (to_wallet) REFERENCES wallets(address)
);

CREATE INDEX idx_transactions_from_wallet ON transactions(from_wallet);
CREATE INDEX idx_transactions_to_wallet ON transactions(to_wallet);
CREATE INDEX idx_transactions_created_at ON transactions(created_at); 