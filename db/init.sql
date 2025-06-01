CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    stripe_customer_id VARCHAR(100),
    password TEXT NOT NULL
);

CREATE TABLE credit_cards (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    stripe_payment_method_id VARCHAR(100) UNIQUE NOT NULL,
    brand VARCHAR(50),
    last4 CHAR(4),
    exp_month INT,
    exp_year INT,
);

CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  price_cents INT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE purchases (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  product_id INT REFERENCES products(id) ON DELETE SET NULL,
  quantity INT NOT NULL,
  total_price_cents INT NOT NULL,
  stripe_payment_intent_id VARCHAR(100) UNIQUE NOT NULL,
  purchased_at TIMESTAMP DEFAULT NOW()
);


CREATE TABLE admins (
    user_id INT PRIMARY KEY REFERENCES users(id)
);