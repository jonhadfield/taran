package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListResponse[T any] struct {
	Data  []T `json:"data"`
	Total int `json:"total"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg})
}

// maxJSONBody limits the request body to 1MB for JSON API endpoints.
const maxJSONBodySize = 1 << 20 // 1 MB

func LimitedJSONDecoder(r *http.Request) *json.Decoder {
	r.Body = http.MaxBytesReader(nil, r.Body, maxJSONBodySize)
	return json.NewDecoder(r.Body)
}

// queryInt reads an integer query parameter, falling back when it is absent
// or not a number.
func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
