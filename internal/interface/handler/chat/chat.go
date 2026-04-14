package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	appcmd "github.com/ImaSerix/go-message-service/internal/application/command"
	appqry "github.com/ImaSerix/go-message-service/internal/application/query"
	domainchat "github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/interface/handler"
	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
)

type ChatHandler struct {
	cmd *appcmd.CommandService
	qry *appqry.QueryService
}

func New(cmd *appcmd.CommandService, qry *appqry.QueryService) *ChatHandler {
	return &ChatHandler{cmd: cmd, qry: qry}
}

func (ch *ChatHandler) GetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserID(r.Context())
	if !ok {
		slog.Error("reading user from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	chats, err := ch.qry.HandleGetUserChats(r.Context(), &appqry.GetUserChatsQry{UserID: userID})
	if err != nil {
		handler.LogError(r, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out := GetUserResponse{Chats: []GetUserResponseChat{}}
	for _, c := range chats {
		var lastMessage *GetUserResponseLastMessage
		if c.LastMessageID != nil && c.LastMessageSenderID != nil && c.LastMessageContent != nil && c.LastMessageCreatedAt != nil {
			lastMessage = &GetUserResponseLastMessage{
				ID:        *c.LastMessageID,
				SenderID:  *c.LastMessageSenderID,
				Content:   *c.LastMessageContent,
				CreatedAt: *c.LastMessageCreatedAt,
			}
		}
		// TODO: Тут рассинхрон с README
		out.Chats = append(out.Chats, GetUserResponseChat{
			ID:          c.ID,
			Name:        c.Name,
			LastMessage: lastMessage,
			UnreadCount: c.UnreadCount,
		})
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out)
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
		return
	}

	id, err := ch.cmd.HandleCreateChat(r.Context(), &appcmd.CreateChatCmd{
		CallerID:     userID,
		Name:         req.Name,
		Participants: req.Participants,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainchat.ErrInvalidChatName):
			http.Error(w, "invalid name", http.StatusBadRequest)
		case errors.Is(err, domainchat.ErrCallerIsNotParticipant):
			http.Error(w, "user is not participant", http.StatusForbidden)
		case errors.Is(err, appcmd.ErrAlreadyExists):
			http.Error(w, "chat already exists", http.StatusConflict)
		default:
			var missing *appcmd.ErrUsersNotExist
			if errors.As(err, &missing) {
				http.Error(w, fmt.Sprintf("not existing participants %v", missing.Users), http.StatusBadRequest)
				return
			}
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateChatResponse{ID: id})
}
