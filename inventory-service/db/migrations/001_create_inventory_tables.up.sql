-- Create orders table

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create index on product SKU
CREATE INDEX idx_products_sku ON products(sku);
CREATE INDEX idx_products_category ON products(category);


-- Create inventory_items table
CREATE TABLE IF NOT EXISTS inventory_items (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    reserved_quantity INTEGER NOT NULL,
    available_quantity INTEGER NOT NULL,
    reorder_point INTEGER NOT NULL,
    reorder_quantity INTEGER NOT NULL,
    stock_status TEXT NOT NULL,
    location_code TEXT NOT NULL,
    last_stocked_at TIMESTAMP NOT NULL, 
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create indices for inventory items
CREATE INDEX idx_inventory_items_product_id ON inventory_items(product_id);
CREATE INDEX idx_inventory_items_stock_status ON inventory_items(stock_status);
CREATE INDEX idx_inventory_items_location_code ON inventory_items(location_code);

-- Create inventory_transactions table
CREATE TABLE IF NOT EXISTS inventory_transactions (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL,
    type TEXT NOT NULL,
    reference_id TEXT,
    note TEXT,
    performed_by TEXT NOT NULL,
    transacted_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

-- Create indices for inventory_transactions
CREATE INDEX idx_inventory_transactions_product_id ON inventory_transactions(product_id);
CREATE INDEX idx_inventory_transactions_type ON inventory_transactions(type);
CREATE INDEX idx_inventory_transactions_reference_id ON inventory_transactions(reference_id);
CREATE INDEX idx_inventory_transactions_transacted_at ON inventory_transactions(transacted_at);
