package rest

import (
	"encoding/json"
	"net/http"

	"shortify/server/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal error"

	switch err {
	case domain.ErrInvalidInput:
		status = http.StatusBadRequest
		message = "invalid input"
	case domain.ErrInvalidEmail:
		status = http.StatusBadRequest
		message = "invalid email format"
	case domain.ErrPasswordTooShort:
		status = http.StatusBadRequest
		message = "А что ещё у тебя такое же короткое как этот пароль?"
	case domain.ErrProfanity:
		status = http.StatusBadRequest
		message = "Фу как некультурно"
	case domain.ErrInvalidURL:
		status = http.StatusBadRequest
		message = "invalid url"
	case domain.ErrInvalidCreds:
		status = http.StatusUnauthorized
		message = "invalid credentials"
	case domain.ErrForbidden:
		status = http.StatusForbidden
		message = "forbidden"
	case domain.ErrNotFound:
		status = http.StatusNotFound
		message = "not found"
	case domain.ErrAlreadyExists:
		status = http.StatusConflict
		message = "already exists"
	}

	writeJSON(w, status, map[string]string{"error": message})
}
