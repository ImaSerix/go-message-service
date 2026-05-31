-- Drop an existing table 'messages'
DROP INDEX idx_messages_chat_id_id;
DROP INDEX idx_messages_chat_id_id_asc;

ALTER TABLE chats DROP CONSTRAINT fk_last_message;

DROP TABLE messages;