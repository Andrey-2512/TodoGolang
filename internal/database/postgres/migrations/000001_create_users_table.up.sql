CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username CITEXT UNIQUE,
    hash_password TEXT
    
);