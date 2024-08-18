CREATE TYPE transaction_status AS ENUM ('PENDING', 'SUCCESS', 'FAILURE');

CREATE TABLE transactions (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              user_id UUID,
                              amount INT NOT NULL,
                              account_id UUID NOT NULL,
                              currency_code VARCHAR(10) NOT NULL,
                              reference_id UUID NOT NULL,
                              metadata JSONB,
                              status transaction_status NOT NULL,
                              transaction_type VARCHAR(20) NOT NULL,
                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_account_id ON transactions(account_id);
CREATE INDEX idx_transactions_reference_id ON transactions(reference_id);
CREATE INDEX idx_transactions_currency_code ON transactions(currency_code);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_transaction_type ON transactions(transaction_type);

-- Optionally, a composite index if you often query by these fields together
CREATE INDEX idx_transactions_account_id_status ON transactions(account_id, status);
CREATE INDEX idx_transactions_account_id_type ON transactions(account_id, transaction_type);
