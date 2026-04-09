package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ImaSerix/go-message-service/internal/handler"
	"github.com/ImaSerix/go-message-service/internal/usecase/user"
)

type UserHandler struct {
	uuc user.UsersUseCase
}

func New(uuc user.UsersUseCase) *UserHandler {
	return &UserHandler{
		uuc: uuc,
	}
}

func (uh *UserHandler) SaveLocal(w http.ResponseWriter, r *http.Request) {
	var req SaveLocalRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		handler.LogError(r, err)

		http.Error(w, "invalid json body", http.StatusBadRequest)
	}

	in := user.CreateInput{
		ID:       req.UserID,
		Username: req.Username,
	}

	_, err := uh.uuc.Create(r.Context(), &in)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidUUID):
			http.Error(w, "invalid UUID", http.StatusBadRequest)
		case errors.Is(err, user.ErrInvalidUsername):
			http.Error(w, "invalid username", http.StatusBadRequest)
		case errors.Is(err, user.ErrAlreadyExists):
			http.Error(w, "userID taken", http.StatusConflict)
		default:
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("user created")
}
