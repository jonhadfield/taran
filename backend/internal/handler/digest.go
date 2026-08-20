package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/digest"
	"github.com/hadfielj/taran/backend/internal/domain"
	"github.com/hadfielj/taran/backend/internal/mailer"
)

const shareTokenExpiry = 30 * 24 * time.Hour

type DigestHandler struct {
	Digests           database.DigestRepository
	DigestFeedback    database.DigestFeedbackRepository
	Generator         *digest.Generator
	Preferences       database.PreferenceRepository
	Mailer            mailer.Mailer
	Sessions          database.SessionRepository
	BaseURL           string
	UnsubscribeSecret string
}

func (h *DigestHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	opts := domain.ListOptions{
		Limit:  clampInt(intParam(r, "limit", 50), 1, 200),
		Offset: clampInt(intParam(r, "offset", 0), 0, 100000),
	}

	digests, total, err := h.Digests.List(r.Context(), userID, opts)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to list digests")
		return
	}

	WriteJSON(w, http.StatusOK, ListResponse[domain.Digest]{Data: digests, Total: total})
}

func (h *DigestHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	d, err := h.Digests.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	WriteJSON(w, http.StatusOK, d)
}

// resolvePeriod determines the digest period from the user's frequency preference
// and optional query parameter overrides. Returns an error message suitable for
// the client if validation fails.
func (h *DigestHandler) resolvePeriod(r *http.Request, userID string) (periodType string, periodStart, periodEnd time.Time, errMsg string, statusCode int) {
	periodType = "daily"
	if h.Preferences != nil {
		pref, err := h.Preferences.Get(r.Context(), userID)
		if err == nil && pref.DigestFrequency == "weekly" {
			periodType = "weekly"
		}
	}

	now := time.Now()
	periodEnd = now
	periodStart = now.Add(-24 * time.Hour)
	if periodType == "weekly" {
		periodStart = now.Add(-7 * 24 * time.Hour)
	}

	if v := r.URL.Query().Get("period_start"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return "", time.Time{}, time.Time{}, "invalid period_start format, use RFC3339", http.StatusBadRequest
		}
		periodStart = t
	}
	if v := r.URL.Query().Get("period_end"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return "", time.Time{}, time.Time{}, "invalid period_end format, use RFC3339", http.StatusBadRequest
		}
		periodEnd = t
	}

	if !periodEnd.After(periodStart) {
		return "", time.Time{}, time.Time{}, "period_end must be after period_start", http.StatusBadRequest
	}
	if periodEnd.Sub(periodStart) > 30*24*time.Hour {
		return "", time.Time{}, time.Time{}, "date range cannot exceed 30 days", http.StatusBadRequest
	}

	return periodType, periodStart, periodEnd, "", 0
}

func (h *DigestHandler) Preview(w http.ResponseWriter, r *http.Request) {
	if h.Generator == nil {
		WriteError(w, http.StatusInternalServerError, "digest generation not configured")
		return
	}

	userID := auth.UserIDFromContext(r.Context())

	periodType, periodStart, periodEnd, errMsg, statusCode := h.resolvePeriod(r, userID)
	if errMsg != "" {
		WriteError(w, statusCode, errMsg)
		return
	}

	preview, err := h.Generator.PreviewForUser(r.Context(), userID, periodType, periodStart, periodEnd)
	if err != nil {
		slog.Error("failed to preview digest", "userID", userID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to preview digest")
		return
	}

	if preview == nil {
		WriteError(w, http.StatusUnprocessableEntity, "no emails found in the selected period")
		return
	}

	WriteJSON(w, http.StatusOK, preview)
}

func (h *DigestHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if h.Generator == nil {
		WriteError(w, http.StatusInternalServerError, "digest generation not configured")
		return
	}

	userID := auth.UserIDFromContext(r.Context())

	periodType, periodStart, periodEnd, errMsg, statusCode := h.resolvePeriod(r, userID)
	if errMsg != "" {
		WriteError(w, statusCode, errMsg)
		return
	}

	exists, err := h.Digests.ExistsForPeriod(r.Context(), userID, periodStart, periodEnd)
	if err != nil {
		slog.Error("failed to check existing digest", "userID", userID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to check existing digest")
		return
	}
	if exists {
		WriteError(w, http.StatusConflict, "a digest already exists for this period")
		return
	}

	d, err := h.Generator.GenerateForUser(r.Context(), userID, periodType, periodStart, periodEnd)
	if err != nil {
		slog.Error("failed to generate digest", "userID", userID, "period", periodType, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to generate digest")
		return
	}

	if d == nil {
		WriteError(w, http.StatusUnprocessableEntity, "no emails found in the selected period")
		return
	}

	WriteJSON(w, http.StatusOK, d)
}

func (h *DigestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	_, err := h.Digests.GetByID(r.Context(), userID, id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	if err := h.Digests.Delete(r.Context(), userID, id); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete digest")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DigestHandler) Share(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	token, err := domain.GenerateShareToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	if err := h.Digests.SetShareToken(r.Context(), id, userID, token); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to share digest")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"ShareToken": token})
}

func (h *DigestHandler) Unshare(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	if err := h.Digests.ClearShareToken(r.Context(), id, userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to unshare digest")
		return
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *DigestHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	if h.Mailer == nil {
		WriteError(w, http.StatusServiceUnavailable, "email delivery not configured")
		return
	}

	id := r.PathValue("id")

	d, err := h.Digests.GetByIDInternal(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	email, err := h.Sessions.GetUserEmail(r.Context(), d.UserID)
	if err != nil {
		slog.Error("failed to get user email for digest send", "userID", d.UserID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to look up user email")
		return
	}

	// Auto-generate share token if missing
	if d.ShareToken == nil {
		token, err := domain.GenerateShareToken()
		if err != nil {
			slog.Error("failed to generate share token", "digestID", d.ID, "error", err)
		} else {
			if err := h.Digests.SetShareToken(r.Context(), d.ID, d.UserID, token); err != nil {
				slog.Error("failed to set share token", "digestID", d.ID, "error", err)
			} else {
				d.ShareToken = &token
			}
		}
	}

	// Populate transient EmailSummaries from persisted Items for email rendering
	for _, item := range d.Items {
		d.EmailSummaries = append(d.EmailSummaries, domain.DigestEmailSummary{
			EmailID:    item.EmailID,
			Subject:    item.Subject,
			SenderName: item.FromName,
			Summary:    item.Summary,
		})
	}

	var unsubURL string
	if h.BaseURL != "" && h.UnsubscribeSecret != "" {
		unsubURL = mailer.GenerateUnsubscribeURL(h.BaseURL, d.UserID, h.UnsubscribeSecret)
	}

	if err := h.Mailer.SendDigest(r.Context(), email, "", d, unsubURL); err != nil {
		slog.Error("failed to send digest email", "digestID", d.ID, "error", err)
		WriteError(w, http.StatusInternalServerError, "failed to send digest email")
		return
	}

	now := time.Now()
	if err := h.Digests.SetSentAt(r.Context(), d.ID, now); err != nil {
		slog.Error("failed to set digest sent_at", "digestID", d.ID, "error", err)
	}

	WriteJSON(w, http.StatusOK, map[string]string{"status": "sent", "to": email})
}

func (h *DigestHandler) GetPublic(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		WriteError(w, http.StatusBadRequest, "missing token")
		return
	}

	d, err := h.Digests.GetByShareToken(r.Context(), token)
	if err != nil || d == nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	// Shared links expire 30 days after being shared. Measuring from the
	// digest's own creation time meant sharing an older digest handed out a
	// link that was already expired.
	sharedAt := d.CreatedAt
	if d.ShareTokenCreatedAt != nil {
		sharedAt = *d.ShareTokenCreatedAt
	}
	if time.Since(sharedAt) > shareTokenExpiry {
		WriteError(w, http.StatusGone, "this shared digest has expired")
		return
	}

	// Strip sensitive fields for public view
	d.UserID = ""

	WriteJSON(w, http.StatusOK, d)
}

func (h *DigestHandler) UpsertFeedback(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	digestID := r.PathValue("id")

	// Verify the digest belongs to this user
	if _, err := h.Digests.GetByID(r.Context(), userID, digestID); err != nil {
		WriteError(w, http.StatusNotFound, "digest not found")
		return
	}

	var body struct {
		Rating string `json:"Rating"`
	}
	if err := LimitedJSONDecoder(r).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !domain.IsValidRating(body.Rating) {
		WriteError(w, http.StatusBadRequest, "rating must be 'useful' or 'not_useful'")
		return
	}

	fb := &domain.DigestFeedback{
		ID:        uuid.New().String(),
		UserID:    userID,
		DigestID:  digestID,
		Rating:    body.Rating,
		CreatedAt: time.Now(),
	}

	if err := h.DigestFeedback.Upsert(r.Context(), fb); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to save feedback")
		return
	}

	WriteJSON(w, http.StatusOK, fb)
}

func (h *DigestHandler) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	digestID := r.PathValue("id")

	if err := h.DigestFeedback.Delete(r.Context(), userID, digestID); err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to delete feedback")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DigestHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	digestID := r.PathValue("id")

	fb, err := h.DigestFeedback.GetByDigestID(r.Context(), userID, digestID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "failed to get feedback")
		return
	}
	if fb == nil {
		WriteJSON(w, http.StatusOK, map[string]any{"Rating": nil})
		return
	}

	WriteJSON(w, http.StatusOK, fb)
}
