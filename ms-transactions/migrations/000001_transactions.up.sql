CREATE TABLE IF NOT EXISTS accounts (
    user_id UUID PRIMARY KEY,
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0.00,
    CHECK (balance >= 0)
);

CREATE TYPE transaction_type AS ENUM ('credit', 'debit');

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES accounts (user_id),
    type transaction_type NOT NULL,
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);