package message

type Content string

const MaxContentLenght = 10000

func NewContent(content string) (Content, error) {
	if len(content) == 0 {
		return "", ErrContentEmpty
	}
	if len(content) > MaxContentLenght {
		return "", ErrContentToLong
	}
	return Content(content), nil
}
