-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNIQUE NOT NULL,
    status INT DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channels (
    name VARCHAR(50)
);

-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id BIGINT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS movies (
    id BIGINT PRIMARY KEY,
    link VARCHAR(100),
    title VARCHAR(255),
    genre VARCHAR(100),
    release_year INT
);

