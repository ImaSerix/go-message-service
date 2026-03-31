-- Create a new table 'chats' with a primary key and columns
CREATE TABLE chats (
    id UUID PRIMARY KEY,
    name varchar(32),
    created_at TIMESTAMP DEFAULT now()
);