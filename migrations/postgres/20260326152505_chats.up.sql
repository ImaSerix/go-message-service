-- Create a new table 'chats' with a primary key and columns
CREATE TABLE chats (
    id UUID PRIMARY KEY,
    name varchar(32) NOT NULL,
    message_count int DEFAULT 0,
    last_message_id bigint,
    last_message_sender_id UUID REFERENCES users(id),
    last_message_content text,
    last_message_created_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT now()
);