package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/mailer"
)

type PreferenceHandler struct {
	Preferences       database.PreferenceRepository
	UnsubscribeSecret string
}

func (h *PreferenceHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	pref, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	WriteJSON(w, http.StatusOK, pref)
}

type updatePreferenceRequest struct {
	DigestEmail     *bool   `json:"DigestEmail"`
	DigestFrequency *string `json:"DigestFrequency"`
	DigestHour      *int    `json:"DigestHour"`
	DigestDay       *int    `json:"DigestDay"`
	DigestTimezone  *string `json:"DigestTimezone"`
	TopicLimit      *int    `json:"TopicLimit"`
	DigestStyle       *string   `json:"DigestStyle"`
	InterestKeywords  *[]string `json:"InterestKeywords"`
	ExclusionKeywords *[]string `json:"ExclusionKeywords"`
}

func (h *PreferenceHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var req updatePreferenceRequest
	if err := LimitedJSONDecoder(r).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate frequency
	if req.DigestFrequency != nil {
		if *req.DigestFrequency != "daily" && *req.DigestFrequency != "weekly" {
			WriteError(w, http.StatusBadRequest, "frequency must be 'daily' or 'weekly'")
			return
		}
	}

	// Validate hour
	if req.DigestHour != nil {
		if *req.DigestHour < 0 || *req.DigestHour > 23 {
			WriteError(w, http.StatusBadRequest, "hour must be between 0 and 23")
			return
		}
	}

	// Validate day of week
	if req.DigestDay != nil {
		if *req.DigestDay < 0 || *req.DigestDay > 6 {
			WriteError(w, http.StatusBadRequest, "day must be between 0 (Sunday) and 6 (Saturday)")
			return
		}
	}

	// Validate timezone
	if req.DigestTimezone != nil {
		if _, err := time.LoadLocation(*req.DigestTimezone); err != nil {
			WriteError(w, http.StatusBadRequest, "invalid timezone")
			return
		}
	}

	// Validate topic limit
	if req.TopicLimit != nil {
		if *req.TopicLimit < 5 || *req.TopicLimit > 50 {
			WriteError(w, http.StatusBadRequest, "topic limit must be between 5 and 50")
			return
		}
	}

	// Validate digest style
	if req.DigestStyle != nil {
		if *req.DigestStyle != "concise" && *req.DigestStyle != "detailed" {
			WriteError(w, http.StatusBadRequest, "digest style must be 'concise' or 'detailed'")
			return
		}
	}

	// Validate keywords
	if req.InterestKeywords != nil {
		cleaned, err := validateKeywords(*req.InterestKeywords)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.InterestKeywords = &cleaned
	}
	if req.ExclusionKeywords != nil {
		cleaned, err := validateKeywords(*req.ExclusionKeywords)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		req.ExclusionKeywords = &cleaned
	}

	// Load existing preference and merge
	existing, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	if req.DigestEmail != nil {
		existing.DigestEmail = *req.DigestEmail
	}
	if req.DigestFrequency != nil {
		existing.DigestFrequency = *req.DigestFrequency
	}
	if req.DigestHour != nil {
		existing.DigestHour = *req.DigestHour
	}
	if req.DigestDay != nil {
		existing.DigestDay = *req.DigestDay
	}
	if req.DigestTimezone != nil {
		existing.DigestTimezone = *req.DigestTimezone
	}
	if req.TopicLimit != nil {
		existing.TopicLimit = *req.TopicLimit
	}
	if req.DigestStyle != nil {
		existing.DigestStyle = *req.DigestStyle
	}
	if req.InterestKeywords != nil {
		existing.InterestKeywords = *req.InterestKeywords
	}
	if req.ExclusionKeywords != nil {
		existing.ExclusionKeywords = *req.ExclusionKeywords
	}

	if err := h.Preferences.Upsert(r.Context(), existing); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to update preferences")
		return
	}

	updated, err := h.Preferences.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	WriteJSON(w, http.StatusOK, updated)
}

// Unsubscribe handles the public one-click unsubscribe endpoint (no auth required).
func (h *PreferenceHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	token := r.URL.Query().Get("token")

	if uid == "" || token == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, unsubscribeHTML("Invalid unsubscribe link."))
		return
	}

	if !mailer.ValidateUnsubscribeToken(uid, token, h.UnsubscribeSecret) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, unsubscribeHTML("Invalid or expired unsubscribe link."))
		return
	}

	pref, err := h.Preferences.Get(r.Context(), uid)
	if err != nil {
		slog.Error("unsubscribe: failed to get preferences", "userID", uid, "error", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, unsubscribeHTML("Something went wrong. Please try again later."))
		return
	}

	pref.DigestEmail = false
	if err := h.Preferences.Upsert(r.Context(), pref); err != nil {
		slog.Error("unsubscribe: failed to update preferences", "userID", uid, "error", err)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, unsubscribeHTML("Something went wrong. Please try again later."))
		return
	}

	slog.Info("user unsubscribed via email link", "userID", uid)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, unsubscribeHTML(`You have been unsubscribed from MailBrief digest emails. You can re-enable them anytime in your <a href="https://mailbrief.io/settings" style="color:#0066cc;text-decoration:underline;">settings</a>.`))
}

var keywordAllowedPattern = regexp.MustCompile(`^[a-zA-Z0-9 \-\.\/\+]+$`)

func validateKeywords(keywords []string) ([]string, error) {
	if len(keywords) > 20 {
		return nil, fmt.Errorf("maximum 20 keywords allowed")
	}
	seen := make(map[string]bool, len(keywords))
	cleaned := make([]string, 0, len(keywords))
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}
		if len(kw) > 30 {
			return nil, fmt.Errorf("keyword %q exceeds 30 character limit", kw)
		}
		if !keywordAllowedPattern.MatchString(kw) {
			return nil, fmt.Errorf("keyword %q contains invalid characters", kw)
		}
		lower := strings.ToLower(kw)
		if !seen[lower] {
			seen[lower] = true
			cleaned = append(cleaned, kw)
		}
	}
	return cleaned, nil
}

func unsubscribeHTML(message string) string {
	return `<!DOCTYPE html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Unsubscribe - MailBrief</title></head>` +
		`<body style="font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;display:flex;justify-content:center;align-items:center;min-height:100vh;margin:0;background:#fafafa;">` +
		`<div style="text-align:center;max-width:400px;padding:40px 20px;">` +
		`<h1 style="font-size:20px;margin-bottom:12px;">MailBrief</h1>` +
		`<p style="color:#555;line-height:1.5;">` + message + `</p>` +
		`</div></body></html>`
}
