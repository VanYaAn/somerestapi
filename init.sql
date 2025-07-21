CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS ads (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL CHECK (char_length(title) >= 3),
    description TEXT NOT NULL CHECK (char_length(description) >= 10),
    image_url VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2)  NOT NULL CHECK (price >= 0),
    user_id INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);