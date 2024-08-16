CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       email VARCHAR(255) UNIQUE NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       first_name VARCHAR(100) NOT NULL,
                       last_name VARCHAR(100) NOT NULL,
                       phone_number VARCHAR(20),
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
                       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_users_id ON users(id);

