package message

import "time"

type Message struct {
	id        SnowflakeID
	chatID    ChatID
	senderID  SenderID
	content   Content
	createdAt *time.Time
}

func NewMessage(id SnowflakeID, chatID ChatID, senderID SenderID, content Content) *Message {
	return &Message{
		id:       id,
		chatID:   chatID,
		senderID: senderID,
		content:  content,
	}
}

func (m Message) ID() SnowflakeID {
	return m.id
}

func (m Message) ChatID() ChatID {
	return m.chatID
}

func (m Message) SenderID() SenderID {
	return m.senderID
}

func (m Message) Content() Content {
	return m.content
}

func (m Message) CreatedAt() *time.Time {
	return m.createdAt
}
