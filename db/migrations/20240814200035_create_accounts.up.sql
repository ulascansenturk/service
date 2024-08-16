CREATE TYPE account_status AS ENUM ('ACTIVE', 'BLACKLISTED', 'CLOSED');

-- Create the accounts table using the enum type for the status column
CREATE TABLE accounts (
                          id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                          user_id UUID NOT NULL,
                          balance BIGINT NOT NULL CHECK (balance >= 0),
                          currency VARCHAR(3) NOT NULL,
                          status account_status NOT NULL DEFAULT 'ACTIVE',  -- Using the enum type
                          created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_accounts_user_id ON accounts(user_id);
CREATE INDEX idx_accounts_status ON accounts(status);
CREATE INDEX idx_accounts_currency ON accounts(currency);
