-- Example schema for dbx package demonstration
-- Run this in your PostgreSQL database before running the example

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create customers table
CREATE TABLE IF NOT EXISTS customer (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create invoice table
CREATE TABLE IF NOT EXISTS invoice (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customer(id),
    amount DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert sample data
INSERT INTO users (name, email, active) VALUES
    ('John Doe', 'john@example.com', true),
    ('Jane Smith', 'jane@example.com', true),
    ('Bob Johnson', 'bob@example.com', false),
    ('Alice Brown', 'alice@example.com', true),
    ('Charlie Wilson', 'charlie@example.com', true)
ON CONFLICT (email) DO NOTHING;

INSERT INTO customer (name, email) VALUES
    ('Acme Corp', 'acme@example.com'),
    ('TechStart Inc', 'tech@example.com'),
    ('Global Solutions', 'global@example.com'),
    ('Local Business', 'local@example.com')
ON CONFLICT (email) DO NOTHING;

INSERT INTO invoice (customer_id, amount) VALUES
    (1, 1500.00),
    (1, 750.50),
    (2, 2200.00),
    (3, 500.00),
    (4, 1200.00),
    (2, 3000.00)
ON CONFLICT DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_invoice_customer_id ON invoice(customer_id);
CREATE INDEX IF NOT EXISTS idx_invoice_amount ON invoice(amount); 