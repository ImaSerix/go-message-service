package chat_test

import (
	"testing"

	"github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/domain/message"
	"github.com/google/uuid"
)

func TestName(t *testing.T) {
	tests := []struct {
		name        string
		chatName    string
		expChatName string
		expErr      error
	}{
		{name: "success", chatName: "niceName", expChatName: "niceName", expErr: nil},
		{name: "short name", chatName: "ab", expChatName: "ab", expErr: nil},
		{name: "long name", chatName: "longlonglonglonglonglonglongname", expChatName: "longlonglonglonglonglonglongname", expErr: nil},
		{name: "long name with spaces", chatName: "     longlonglonglonglonglonglongname     ", expChatName: "longlonglonglonglonglonglongname", expErr: nil},
		{name: "with spaces name", chatName: " name    ", expChatName: "name", expErr: nil},

		{name: "empty name", chatName: "", expErr: chat.ErrInvalidChatName},
		{name: "very short name", chatName: "a", expErr: chat.ErrInvalidChatName},
		{name: "very long name", chatName: "vlonglonglonglonglonglonglongname", expErr: chat.ErrInvalidChatName},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			n, err := chat.NewName(test.chatName)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if string(n) != test.expChatName {
				t.Fatalf("expected %s, got %s", test.expChatName, n)
			}
		})
	}
}

//TODO: Я всё таки считаю, что чат с "самим собой" можно будет

func TestNewChat(t *testing.T) {

	callerID := uuid.New()
	tests := []struct {
		name         string
		chatID       uuid.UUID
		callerID     uuid.UUID
		participants []uuid.UUID
		expErr       error
	}{
		{name: "success", chatID: uuid.New(), callerID: callerID, participants: []uuid.UUID{callerID, uuid.New(), uuid.New(), uuid.New()}, expErr: nil},
		{name: "empty participants", chatID: uuid.New(), callerID: uuid.New(), participants: []uuid.UUID{}, expErr: chat.ErrParticipantsEmpty},
		{name: "caller not participant", chatID: uuid.New(), callerID: uuid.New(), participants: []uuid.UUID{uuid.New(), uuid.New()}, expErr: chat.ErrCallerIsNotParticipant},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id := chat.ChatID(test.chatID)
			callerID := chat.ParticipantID(test.callerID)
			name, _ := chat.NewName("test")
			participants := []chat.ParticipantID{}
			for _, p := range test.participants {
				participants = append(participants, chat.ParticipantID(p))
			}
			c, err := chat.NewChat(id, name, callerID, participants)
			if err != test.expErr {
				t.Fatalf("expected %v, got %v", test.expErr, err)
			}
			if err != nil && c != nil {
				t.Fatalf("got chat, but error occurred")
			}
			if err == nil && c == nil {
				t.Fatalf("got nil chat, but no error")
			}
			if c != nil && c.ID() != id {
				t.Fatal("id mismatch")
			}
			if c != nil && c.Name() != name {
				t.Fatal("name mismatch")
			}
		})
	}
}

func TestSendMessage(t *testing.T) {

	senderID := uuid.New()
	tests := []struct {
		name         string
		chatID       uuid.UUID
		senderID     uuid.UUID
		participants []uuid.UUID
		expErr       error
	}{
		{name: "success", chatID: uuid.New(), senderID: senderID, participants: []uuid.UUID{senderID, uuid.New(), uuid.New(), uuid.New()}, expErr: nil},
		{name: "sender not participant", chatID: uuid.New(), senderID: uuid.New(), participants: []uuid.UUID{uuid.New(), uuid.New()}, expErr: chat.ErrCallerIsNotParticipant},
	}

	for _, test := range tests {
		senderID := chat.ParticipantID(test.senderID)
		participants := []chat.ParticipantID{}
		for _, p := range test.participants {
			participants = append(participants, chat.ParticipantID(p))
		}
		c := chat.RehydrateChat(chat.ChatID(test.chatID), chat.Name(test.name), participants, nil, nil)

		testMessage := message.NewMessage(
			message.SnowflakeID(123456),
			message.ChatID(test.chatID),
			message.SenderID(test.senderID),
			message.Content("test content"),
		)

		err := c.SendMessage(senderID, testMessage)
		if err != test.expErr {
			t.Fatalf("expected %v, got %v", test.expErr, err)
		}
	}
}

func existsInChatMaker(exists bool) func(chat.SnowflakeID, chat.ChatID) (bool, error) {
	return func(chat.SnowflakeID, chat.ChatID) (bool, error) {
		return exists, nil
	}
}
func TestReadMessage(t *testing.T) {

	type userType struct {
		id                uuid.UUID
		lastReadMessageID int64
	}
	type chatType struct {
		participants []uuid.UUID
	}

	readerID := uuid.New()
	tests := []struct {
		name         string
		user         userType
		chat         chatType
		readMessage  int64
		existsInChat bool
		expErr       error
	}{
		{
			name: "success",
			user: userType{
				id:                readerID,
				lastReadMessageID: 10001,
			},
			chat: chatType{
				participants: []uuid.UUID{readerID, uuid.New()},
			},
			readMessage:  10002,
			existsInChat: true,
			expErr:       nil,
		},
		{
			name: "not participant",
			user: userType{
				id:                readerID,
				lastReadMessageID: 10001,
			},
			chat: chatType{
				participants: []uuid.UUID{uuid.New(), uuid.New()},
			},
			expErr: chat.ErrReaderIsNotParticipant,
		},
		{
			name: "not exists in chat",
			user: userType{
				id:                readerID,
				lastReadMessageID: 10001,
			},
			chat: chatType{
				participants: []uuid.UUID{readerID, uuid.New()},
			},
			existsInChat: false,
			expErr:       chat.ErrMessageNotInChat,
		},
		{
			name: "newer message read",
			user: userType{
				id:                readerID,
				lastReadMessageID: 10002,
			},
			chat: chatType{
				participants: []uuid.UUID{readerID, uuid.New()},
			},
			readMessage:  10001,
			existsInChat: true,
			expErr:       chat.ErrNewerMessageIsAlreadyRead,
		},
		{
			name: "newer message read",
			user: userType{
				id:                readerID,
				lastReadMessageID: 10002,
			},
			chat: chatType{
				participants: []uuid.UUID{readerID, uuid.New()},
			},
			readMessage:  10001,
			existsInChat: true,
			expErr:       chat.ErrNewerMessageIsAlreadyRead,
		},
		{
			name: "non message read (before)",
			user: userType{
				id:                readerID,
				lastReadMessageID: 0,
			},
			chat: chatType{
				participants: []uuid.UUID{readerID, uuid.New()},
			},
			readMessage:  10001,
			existsInChat: true,
			expErr:       nil,
		},
	}

	for _, test := range tests {
		readerID := chat.ParticipantID(test.user.id)
		readMessageID := chat.SnowflakeID(test.user.lastReadMessageID)
		participants := []chat.ParticipantID{}
		for _, p := range test.chat.participants {
			participants = append(participants, chat.ParticipantID(p))
		}

		lastReadedMessage := map[chat.ParticipantID]*chat.SnowflakeID{}
		if readMessageID == 0 {
			lastReadedMessage[readerID] = nil
		} else {
			lastReadedMessage[readerID] = &readMessageID
		}

		c := chat.RehydrateChat(chat.ChatID(uuid.New()), chat.Name("test"), participants, nil, lastReadedMessage)

		err := c.ReadMessage(readerID, chat.SnowflakeID(test.readMessage), existsInChatMaker(test.existsInChat))
		if err != test.expErr {
			t.Fatalf("expected %v, got %v", test.expErr, err)
		}
		if id, _ := c.LastReadMessageID(readerID); err == nil && *id != chat.SnowflakeID(test.readMessage) {
			t.Fatalf("lastReadMessage mismatch")
		}
		if id, _ := c.LastReadMessageID(readerID); err != nil && *id != readMessageID {
			t.Fatalf("readed message changed, when error")
		}
	}
}
