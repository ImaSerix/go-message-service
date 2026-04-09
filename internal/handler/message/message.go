package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/domain"
	"github.com/ImaSerix/go-message-service/internal/handler"
	"github.com/ImaSerix/go-message-service/internal/handler/auth"
	"github.com/ImaSerix/go-message-service/internal/usecase/message"
)

type MessageHandler struct {
	muc message.MessagesUseCase
}

func New(muc message.MessagesUseCase) *MessageHandler {
	return &MessageHandler{
		muc: muc,
	}
}

// NOTE: требует запустить middleware, которое обработает все параметры и засунет их в контекст, вернув ошибки, если их нет
func (muc *MessageHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	chatID, ok := ChatID(r.Context())
	if !ok {
		slog.Error("reading chat from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var cursor *int64
	// NOTE: Получаем более новые сообщения первыми, далее уже сатые по дефолту
	// TODO: Описать это поведение ещё в доках, проверить не налажал ли я тут по итогу
	direction := domain.DirectionForward
	afterParam, ok := AfterParam(r.Context())
	if ok {
		cursor = &afterParam
		direction = domain.DirectionForward
	}
	BeforeParam, ok := BeforeParam(r.Context())
	if ok {
		if cursor != nil {
			http.Error(w, "used before and after at the same time", http.StatusInternalServerError)
			return
		}
		direction = domain.DirectionBackward
		cursor = &BeforeParam
	}
	// TODO: Опять же надо придумать ошибку и под ограничение лимита, ибо нефиг
	limitParam, ok := LimitParam(r.Context())
	if !ok {
		slog.Error("reading limit from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	in := message.GetMessagesInput{
		ChatID:    chatID,
		UserID:    userID,
		Cursor:    cursor,
		Direction: direction,
		Limit:     limitParam,
	}

	messages, err := muc.muc.Get(r.Context(), &in)
	if err != nil {
		switch {
		case errors.Is(err, message.ErrUserIsNotParticipant):
			// NOTE: фактически middleware до, должен был обработать этот случай, поэтому handler не берётся обрабатывать этот случай
			handler.LogError(r, errors.New("get message handler, called without authоМiddleware"))
			http.Error(w, "internal error", http.StatusInternalServerError)
		case errors.Is(err, message.ErrChatNotExists):
			http.Error(w, fmt.Sprintf("invalid chat %v", in.ChatID), http.StatusBadRequest)
		case errors.Is(err, message.ErrMessageNotInChat):
			http.Error(w, fmt.Sprintf("no message %v in chat %v", *in.Cursor, in.ChatID), http.StatusBadRequest)
		// TODO: Вот это нужно добавить и в readme и в obsidian
		case errors.Is(err, message.ErrInvalidLimit):
			http.Error(w, "invalid limit", http.StatusBadRequest)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	out := GetMessagesResponse{
		Messages:     []GetMessagesResponseMessage{},
		CursorBefore: cursor,
		CursorAfter:  messages.NewCursor,
	}
	for _, mDTO := range messages.Messages {
		m := GetMessagesResponseMessage{
			ID:        mDTO.ID,
			SenderID:  mDTO.SenderID,
			Content:   mDTO.Content,
			CreatedAt: mDTO.CreatedAt,
		}
		out.Messages = append(out.Messages, m)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}

// TODO: Надо в целом проверить на соответсвие, иначе тупо получается
func (muc *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	chatID, ok := ChatID(r.Context())
	if !ok {
		slog.Error("reading chat from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req SendMessageRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)

		http.Error(w, "invalid json body", http.StatusBadRequest)
	}

	in := message.CreateMessageInput{
		ChatID:  chatID,
		UserID:  userID,
		Content: req.Content,
	}

	mes, err := muc.muc.Create(r.Context(), &in)
	if err != nil {
		// TODO: Опять же чего-то в описании не хватает, именно ошибок!!!!!!!
		switch {
		case errors.Is(err, message.ErrUserIsNotParticipant):
			// NOTE: фактически middleware до, должен был обработать этот случай, поэтому handler не берётся обрабатывать этот случай
			handler.LogError(r, errors.New("get message handler, called without authоМiddleware"))
			http.Error(w, "internal error", http.StatusInternalServerError)
		case errors.Is(err, message.ErrChatNotExists):
			http.Error(w, fmt.Sprintf("invalid chat %v", in.ChatID), http.StatusBadRequest)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
	}

	out := SendMessageResponse{
		ID: mes.ID,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(out)
}

func (muc *MessageHandler) Read(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// TODO: Опять же надо подправить
	chatID, ok := ChatID(r.Context())
	if !ok {
		slog.Error("reading chat from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req ReadMessageRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)

		http.Error(w, "invalid json body", http.StatusBadRequest)
	}

	in := message.ReadMessageInput{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: req.LastReadMessageID,
	}

	err := muc.muc.MarkAsRead(r.Context(), &in)
	if err != nil {
		switch {
		case errors.Is(err, message.ErrUserIsNotParticipant):
			// NOTE: фактически middleware до, должен был обработать этот случай, поэтому handler не берётся обрабатывать этот случай
			handler.LogError(r, errors.New("get message handler, called without authоМiddleware"))
			http.Error(w, "internal error", http.StatusInternalServerError)
		case errors.Is(err, message.ErrChatNotExists):
			http.Error(w, fmt.Sprintf("invalid chat %v", in.ChatID), http.StatusBadRequest)
		case errors.Is(err, message.ErrMessageNotInChat):
			http.Error(w, fmt.Sprintf("no message %v in chat %v", in.MessageID, in.ChatID), http.StatusBadRequest)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
