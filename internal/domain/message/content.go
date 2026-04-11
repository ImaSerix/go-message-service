package message

type Content string

func NewContent(content string) (Content, error) {
	if len(content) == 0 {
		return "", ErrContentEmpty
	}
	return Content(content), nil
}
