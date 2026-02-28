CREATE TABLE tax_rates (
    id SERIAL PRIMARY KEY,
    jurisdiction_type VARCHAR(50) NOT NULL,
    jurisdiction_name VARCHAR(100) NOT NULL,
    composite_rate DECIMAL(7, 5) NOT NULL,
    state_rate DECIMAL(7, 5) NOT NULL,
    county_rate DECIMAL(7, 5) NOT NULL,
    city_rate DECIMAL(7, 5) NOT NULL,
    special_rate DECIMAL(7, 5) NOT NULL,
    special_name VARCHAR(50),
    UNIQUE(jurisdiction_type, jurisdiction_name)
);

CREATE INDEX idx_tax_rates_name ON tax_rates(jurisdiction_name);

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    composite_tax_rate DECIMAL(7, 5) NOT NULL,
    tax_amount DECIMAL(10, 2) NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    breakdown JSONB NOT NULL,
    jurisdictions JSONB NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL
);