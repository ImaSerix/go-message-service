package chat

import (
	"github.com/ImaSerix/go-message-service/internal/domain/message"
)

type Chat struct {
	id   ChatID
	name Name

	participants        []ParticipantID
	lastMessage         *message.Message
	lastReadedMessageID SnowflakeID
}

func NewChat(id ChatID, name Name, callerID ParticipantID, participants []ParticipantID) (*Chat, error) {
	if len(participants) == 0 {
		return nil, ErrParticipantsEmpty
	}

	set := make(map[ParticipantID]struct{}, len(participants))
	for _, p := range participants {
		set[p] = struct{}{}
	}
	if _, ok := set[callerID]; !ok {
		return nil, ErrCallerIsNotParticipant
	}

	return &Chat{
		id:           id,
		name:         name,
		participants: participants,
	}, nil
}

func (c Chat) Participants() []ParticipantID {
	return c.participants
}

func (c Chat) ID() ChatID {
	return c.id
}

func (c Chat) Name() Name {
	return c.Name()
}
