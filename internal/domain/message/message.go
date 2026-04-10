package message

import "time"

type Message struct {
	id        SnowflakeID
	chatID    ChatID
	senderID  SenderID
	content   Content
	CreatedAt *time.Time
}
