package message

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	appcmd "github.com/ImaSerix/go-message-service/internal/application/command"
	appqry "github.com/ImaSerix/go-message-service/internal/application/query"
	"github.com/ImaSerix/go-message-service/internal/domain"
	domainchat "github.com/ImaSerix/go-message-service/internal/domain/chat"
	"github.com/ImaSerix/go-message-service/internal/handler"
	"github.com/ImaSerix/go-message-service/internal/handler/auth"
)

type MessageHandler struct {
	cmd *appcmd.CommandService
	qry *appqry.QueryService
}

func New(cmd *appcmd.CommandService, qry *appqry.QueryService) *MessageHandler {
	return &MessageHandler{cmd: cmd, qry: qry}
}

func (mh *MessageHandler) Get(w http.ResponseWriter, r *http.Request) {
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
	direction := domain.DirectionForward
	afterParam, ok := AfterParam(r.Context())
	if ok {
		cursor = &afterParam
	}
	beforeParam, ok := BeforeParam(r.Context())
	if ok {
		if cursor != nil {
			http.Error(w, "used before and after at the same time", http.StatusBadRequest)
			return
		}
		direction = domain.DirectionBackward
		cursor = &beforeParam
	}

	limitParam, ok := LimitParam(r.Context())
	if !ok {
		slog.Error("reading limit from context failed")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	out, err := mh.qry.HandleGetUserMessages(r.Context(), &appqry.GetUserMessagesQry{
		ChatID:    chatID,
		UserID:    userID,
		Cursor:    cursor,
		Direction: direction,
		Limit:     limitParam,
	})
	if err != nil {
		switch {
		case errors.Is(err, appqry.ErrMessageNotInChat):
			http.Error(w, fmt.Sprintf("no message %v in chat %v", *cursor, chatID), http.StatusBadRequest)
		case errors.Is(err, appqry.ErrInvalidLimit):
			http.Error(w, "invalid limit", http.StatusBadRequest)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	resp := GetMessagesResponse{Messages: []GetMessagesResponseMessage{}, CursorBefore: cursor, CursorAfter: out.NewCursor}
	for _, m := range out.Messages {
		resp.Messages = append(resp.Messages, GetMessagesResponseMessage{ID: m.ID, SenderID: m.SenderID, Content: m.Content, CreatedAt: m.CreatedAt})
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (mh *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
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
		return
	}

	id, err := mh.cmd.HandleCreateMessage(r.Context(), &appcmd.CreateMessageCmd{ChatID: chatID, SenderID: userID, Content: req.Content})
	if err != nil {
		switch {
		case errors.Is(err, appcmd.ErrChatNotExists):
			http.Error(w, fmt.Sprintf("invalid chat %v", chatID), http.StatusBadRequest)
		case errors.Is(err, domainchat.ErrCallerIsNotParticipant):
			http.Error(w, "unauthorizated action", http.StatusForbidden)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(SendMessageResponse{ID: id})
}

func (mh *MessageHandler) Read(w http.ResponseWriter, r *http.Request) {
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

	var req ReadMessageRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	err := mh.cmd.MarkAsRead(r.Context(), &appcmd.ReadMessageCmd{ChatID: chatID, CallerID: userID, MessageID: req.LastReadMessageID})
	if err != nil {
		switch {
		case errors.Is(err, appcmd.ErrChatNotExists):
			http.Error(w, fmt.Sprintf("invalid chat %v", chatID), http.StatusBadRequest)
		case errors.Is(err, appcmd.ErrMessageNotInChat):
			http.Error(w, fmt.Sprintf("no message %v in chat %v", req.LastReadMessageID, chatID), http.StatusBadRequest)
		case errors.Is(err, domainchat.ErrReaderIsNotParticipant):
			http.Error(w, "unauthorizated action", http.StatusForbidden)
		default:
			handler.LogError(r, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
