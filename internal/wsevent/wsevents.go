package wsevent

type WSEvent interface {
	Name() string
}

type NewChat struct {
	Chat Chat `json:"chat"`
}

func (e *NewChat) Name() string {
	return "new_chat"
}

type NewMessage struct {
	Message Message `json:"message"`
}

func (e *NewMessage) Name() string {
	return "new_message"
}
