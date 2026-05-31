-- Create a new table 'users' with a primary key and columns
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username varchar(32) NOT NULL,
    created_at timestamp DEFAULT now()
);