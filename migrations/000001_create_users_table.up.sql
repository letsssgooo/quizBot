CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(32) NOT NULL,
    full_name VARCHAR(150),
    role VARCHAR(20) NOT NULL,
    student_group VARCHAR,
    created_at TIMESTAMP NOT NULL
);