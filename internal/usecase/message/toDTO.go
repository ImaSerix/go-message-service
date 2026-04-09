package message

import (
	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/dto"
)

func toDTO(m *domain.Message) *dto.Message {
	return &dto.Message{
		ID:        int64(m.ID),
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
}

func toDTOList(messages []*domain.Message) []*dto.Message {
	res := make([]*dto.Message, 0, len(messages))
	for _, m := range messages {
		res = append(res, toDTO(m))
	}
	return res
}
