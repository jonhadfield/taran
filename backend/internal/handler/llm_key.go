package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/crypto"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type LLMKeyHandler struct {
	Keys      database.UserLLMKeyRepository
	Encryptor *crypto.Encryptor
}

type llmKeyResponse struct {
	Provider  string `json:"Provider"`
	KeyHint   string `json:"KeyHint"`
	Model     string `json:"Model"`
	IsActive  bool   `json:"IsActive"`
	CreatedAt string `json:"CreatedAt"`
	UpdatedAt string `json:"UpdatedAt"`
}

func (h *LLMKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	if h.Encryptor == nil {
		WriteError(w, http.StatusNotImplemented, "BYOK not configured")
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	keys, err := h.Keys.ListByUser(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list keys")
		return
	}

	resp := make([]llmKeyResponse, 0, len(keys))
	for _, k := range keys {
		resp = append(resp, llmKeyResponse{
			Provider:  k.Provider,
			KeyHint:   k.KeyHint,
			Model:     k.Model,
			IsActive:  k.IsActive,
			CreatedAt: k.CreatedAt.Format(time.RFC3339),
			UpdatedAt: k.UpdatedAt.Format(time.RFC3339),
		})
	}

	WriteJSON(w, http.StatusOK, ListResponse[llmKeyResponse]{Data: resp, Total: len(resp)})
}

func (h *LLMKeyHandler) Put(w http.ResponseWriter, r *http.Request) {
	if h.Encryptor == nil {
		WriteError(w, http.StatusNotImplemented, "BYOK not configured")
		return
	}

	provider := r.PathValue("provider")
	if provider != "anthropic" && provider != "openai" {
		WriteError(w, http.StatusBadRequest, "provider must be anthropic or openai")
		return
	}

	var body struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	}
	if err := LimitedJSONDecoder(r).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.APIKey == "" {
		WriteError(w, http.StatusBadRequest, "api_key is required")
		return
	}

	ciphertext, nonce, err := h.Encryptor.Encrypt(body.APIKey)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to encrypt key")
		return
	}

	hint := ""
	if len(body.APIKey) >= 4 {
		hint = "..." + body.APIKey[len(body.APIKey)-4:]
	}

	userID := auth.UserIDFromContext(r.Context())
	now := time.Now()
	key := &domain.UserLLMKey{
		ID:           uuid.New().String(),
		UserID:       userID,
		Provider:     provider,
		EncryptedKey: ciphertext,
		KeyNonce:     nonce,
		KeyHint:      hint,
		Model:        body.Model,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.Keys.Upsert(r.Context(), key); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to save key")
		return
	}

	WriteJSON(w, http.StatusOK, llmKeyResponse{
		Provider:  key.Provider,
		KeyHint:   key.KeyHint,
		Model:     key.Model,
		IsActive:  key.IsActive,
		CreatedAt: key.CreatedAt.Format(time.RFC3339),
		UpdatedAt: key.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *LLMKeyHandler) Patch(w http.ResponseWriter, r *http.Request) {
	if h.Encryptor == nil {
		WriteError(w, http.StatusNotImplemented, "BYOK not configured")
		return
	}

	provider := r.PathValue("provider")
	if provider != "anthropic" && provider != "openai" {
		WriteError(w, http.StatusBadRequest, "provider must be anthropic or openai")
		return
	}

	var body struct {
		IsActive *bool   `json:"is_active"`
		Model    *string `json:"model"`
	}
	if err := LimitedJSONDecoder(r).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userID := auth.UserIDFromContext(r.Context())
	if err := h.Keys.Update(r.Context(), userID, provider, body.IsActive, body.Model); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update key")
		return
	}

	key, err := h.Keys.GetByUserAndProvider(r.Context(), userID, provider)
	if err != nil || key == nil {
		WriteError(w, http.StatusNotFound, "key not found")
		return
	}

	WriteJSON(w, http.StatusOK, llmKeyResponse{
		Provider:  key.Provider,
		KeyHint:   key.KeyHint,
		Model:     key.Model,
		IsActive:  key.IsActive,
		CreatedAt: key.CreatedAt.Format(time.RFC3339),
		UpdatedAt: key.UpdatedAt.Format(time.RFC3339),
	})
}

func (h *LLMKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if h.Encryptor == nil {
		WriteError(w, http.StatusNotImplemented, "BYOK not configured")
		return
	}

	provider := r.PathValue("provider")
	userID := auth.UserIDFromContext(r.Context())

	if err := h.Keys.Delete(r.Context(), userID, provider); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
