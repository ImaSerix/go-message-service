package message

import "github.com/ImaSerix/go-message-service/internal/dto"

type GetMessageOutput struct {
	Messages  []*dto.Message
	HasMore   bool
	NewCursor *int64
}
