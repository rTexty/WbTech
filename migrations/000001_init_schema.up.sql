CREATE TABLE IF NOT EXISTS orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255) NOT NULL,
    entry VARCHAR(10),
    delivery_name VARCHAR(255) NOT NULL,
    delivery_phone VARCHAR(20) NOT NULL,
    delivery_zip VARCHAR(20),
    delivery_city VARCHAR(100) NOT NULL,
    delivery_address VARCHAR(500) NOT NULL,
    delivery_region VARCHAR(100),
    delivery_email VARCHAR(255),
    payment_transaction VARCHAR(255) NOT NULL,
    payment_request_id VARCHAR(255),
    payment_currency VARCHAR(10) NOT NULL,
    payment_provider VARCHAR(100) NOT NULL,
    payment_amount INT NOT NULL,
    payment_payment_dt INT NOT NULL,
    payment_bank VARCHAR(100),
    payment_delivery_cost INT NOT NULL,
    payment_goods_total INT NOT NULL,
    payment_custom_fee INT DEFAULT 0,
    locale VARCHAR(10),
    internal_signature VARCHAR(500),
    customer_id VARCHAR(255) NOT NULL,
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP NOT NULL,
    oof_shard VARCHAR(10)
);

CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) NOT NULL REFERENCES orders(order_uid),
    chrt_id INT NOT NULL,
    track_number VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    rid VARCHAR(255) NOT NULL,
    name VARCHAR(500) NOT NULL,
    sale INT,
    size VARCHAR(50),
    total_price INT NOT NULL,
    nm_id INT NOT NULL,
    brand VARCHAR(255),
    status INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);