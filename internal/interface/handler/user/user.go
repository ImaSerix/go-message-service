package user

import (
	"encoding/json"
	"errors"
	"net/http"

	appcmd "github.com/ImaSerix/go-message-service/internal/application/command"
	domainuser "github.com/ImaSerix/go-message-service/internal/domain/user"
	"github.com/ImaSerix/go-message-service/internal/interface/handler"
)

type UserHandler struct {
	cmd *appcmd.CommandService
}

func New(cmd *appcmd.CommandService) *UserHandler {
	return &UserHandler{cmd: cmd}
}

func (uh *UserHandler) SaveLocal(w http.ResponseWriter, r *http.Request) {
	var req SaveLocalRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	_, err := uh.cmd.HandleCreateUser(r.Context(), &appcmd.CreateUserCmd{UserID: &req.UserID, Username: req.Username})
	if err != nil {
		switch {
		case errors.Is(err, domainuser.ErrInvalidUsername):
			http.Error(w, "invalid username", http.StatusBadRequest)
		case errors.Is(err, appcmd.ErrAlreadyExists):
			http.Error(w, "userID taken", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode("user created")
}
