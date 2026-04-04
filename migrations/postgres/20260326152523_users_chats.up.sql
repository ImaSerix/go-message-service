-- Create a new table 'users_chats' with a primary key and columns
CREATE TABLE chat_users (
    chat_id UUID NOT NULL,
    user_id UUID NOT NULL,
    last_read_message_id BIGINT references messages(id),
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX idx_chat_users_user_id 
ON chat_users (user_id);