-- Create a new table 'messages' with a primary key and columns
CREATE TABLE messages (
    id bigint PRIMARY KEY,
    chat_id UUID references chats(id) NOT NULL,
    sender_id UUID references users(id) NOT NULL,
    content text NOT NULL,
    created_at timestamp DEFAULT now()
);

CREATE INDEX idx_messages_chat_id_id 
ON messages (chat_id, id DESC);

CREATE INDEX idx_messages_chat_id_id_asc 
ON messages (chat_id, id);