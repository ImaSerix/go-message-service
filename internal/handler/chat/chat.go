package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/handler"
	"github.com/ImaSerix/go-message-service/internal/handler/auth"
	"github.com/ImaSerix/go-message-service/internal/usecase/chat"
)

type ChatHandler struct {
	cus chat.ChatsUseCase
}

func New(cus chat.ChatsUseCase) *ChatHandler {
	return &ChatHandler{
		cus: cus,
	}
}

func (ch *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	in := chat.GetUserChatsInput{
		UserID: userID,
	}

	chats, err := ch.cus.GetUserChats(r.Context(), &in)
	if err != nil {
		handler.LogError(r, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := GetUserResponse{
		Chats: []GetUserResponseChat{},
	}
	for _, chat := range chats {
		var lastMessage GetUserResponseLastMessage
		if chat.LastMessage != nil {
			lastMessage.ID = chat.LastMessage.ID
			lastMessage.SenderID = chat.LastMessage.SenderID
			lastMessage.Content = chat.LastMessage.Content
			lastMessage.CreatedAt = chat.LastMessage.CreatedAt
		}
		out.Chats = append(out.Chats, GetUserResponseChat{
			ID:           chat.ID,
			Name:         chat.Name,
			Participants: chat.Participants,
			LastMessage:  &lastMessage,
			MessageCount: chat.MessageCount,
			UnreadCount:  chat.UnreadCount,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}

func (ch *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var req CreateChatRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)

		http.Error(w, "invalid json body", http.StatusBadRequest)
	}

	in := chat.CreateChatInput{
		CallerID:     userID,
		Name:         req.Name,
		Participants: req.Participants,
	}

	c, err := ch.cus.Create(r.Context(), &in)
	if err != nil {
		switch {
		case errors.Is(err, chat.ErrInvalidName):
			http.Error(w, "invalid name", http.StatusBadRequest)
		case errors.Is(err, chat.ErrCallerIsNotParticipant):
			http.Error(w, "user is not participant", http.StatusForbidden)
		default:
			if err, ok := err.(*chat.ErrUsersNotExists); ok {
				http.Error(w, fmt.Sprintf("not existing participants %v", err.Users), http.StatusBadRequest)
			}

			slog.Error("reading user from context failed")
			http.Error(w, "intrenal error", http.StatusInternalServerError)
		}
		return
	}

	out := CreateChatResponse{
		ID: c.ID,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(out)
}
