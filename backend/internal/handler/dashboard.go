package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/hadfielj/taran/backend/internal/auth"
	"github.com/hadfielj/taran/backend/internal/database"
	"github.com/hadfielj/taran/backend/internal/domain"
)

type DashboardHandler struct {
	Emails      database.EmailRepository
	Digests     database.DigestRepository
	Extractions database.ExtractionRepository
}

type DashboardResponse struct {
	Emails        []domain.Email          `json:"emails"`
	EmailTotal    int                     `json:"emailTotal"`
	Digests       []domain.Digest         `json:"digests"`
	UnreadCount   int                     `json:"unreadCount"`
	Stats         domain.UserStats        `json:"stats"`
	Processing    domain.ProcessingStats  `json:"processing"`
	WeeklyHistory []domain.WeekCount      `json:"weeklyHistory"`
	TopTopics     []string                `json:"topTopics"`
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	ctx := r.Context()
	now := time.Now()

	var resp DashboardResponse
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	setErr := func(err error) {
		mu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		mu.Unlock()
	}

	// Fetch recent emails
	wg.Add(1)
	go func() {
		defer wg.Done()
		emails, total, err := h.Emails.List(ctx, userID, domain.ListOptions{Limit: 5})
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		resp.Emails = emails
		resp.EmailTotal = total
		mu.Unlock()
	}()

	// Fetch recent digests
	wg.Add(1)
	go func() {
		defer wg.Done()
		digests, _, err := h.Digests.List(ctx, userID, domain.ListOptions{Limit: 3})
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		resp.Digests = digests
		mu.Unlock()
	}()

	// Fetch unread count
	wg.Add(1)
	go func() {
		defer wg.Done()
		isRead := false
		_, total, err := h.Emails.List(ctx, userID, domain.ListOptions{Limit: 1, IsRead: &isRead})
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		resp.UnreadCount = total
		mu.Unlock()
	}()

	// Fetch processing stats
	wg.Add(1)
	go func() {
		defer wg.Done()
		counts, err := h.Emails.CountByStatus(ctx, userID)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		resp.Processing = domain.ProcessingStats{
			PendingCount:    counts[domain.EmailStatusPending],
			ProcessingCount: counts[domain.EmailStatusProcessing],
			FailedCount:     counts[domain.EmailStatusFailed],
		}
		mu.Unlock()
	}()

	// Fetch stats
	wg.Add(1)
	go func() {
		defer wg.Done()
		weekStart := startOfWeek(now)
		lastWeekStart := weekStart.AddDate(0, 0, -7)

		thisWeek, err := h.Emails.CountByPeriod(ctx, userID, weekStart, now)
		if err != nil {
			setErr(err)
			return
		}
		lastWeek, err := h.Emails.CountByPeriod(ctx, userID, lastWeekStart, weekStart)
		if err != nil {
			setErr(err)
			return
		}
		processed := domain.EmailStatusProcessed
		_, total, err := h.Emails.List(ctx, userID, domain.ListOptions{Limit: 1, Status: &processed})
		if err != nil {
			setErr(err)
			return
		}
		topSenders, err := h.Emails.TopSenders(ctx, userID, weekStart, now, 5)
		if err != nil {
			setErr(err)
			return
		}

		mu.Lock()
		resp.Stats = domain.UserStats{
			EmailsThisWeek: thisWeek,
			EmailsLastWeek: lastWeek,
			TotalEmails:    total,
			TopSenders:     topSenders,
		}
		mu.Unlock()
	}()

	// Fetch weekly email history (8 weeks)
	wg.Add(1)
	go func() {
		defer wg.Done()
		counts, err := h.Emails.CountByWeek(ctx, userID, 8)
		if err != nil {
			setErr(err)
			return
		}
		mu.Lock()
		resp.WeeklyHistory = counts
		mu.Unlock()
	}()

	// Fetch top topics
	if h.Extractions != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			topics, err := h.Extractions.ListTopicsByUser(ctx, userID, 10)
			if err != nil {
				// Non-fatal — topics are optional
				return
			}
			mu.Lock()
			resp.TopTopics = topics
			mu.Unlock()
		}()
	}

	wg.Wait()

	if firstErr != nil {
		WriteError(w, http.StatusInternalServerError, "failed to load dashboard")
		return
	}

	if resp.Emails == nil {
		resp.Emails = []domain.Email{}
	}
	if resp.Digests == nil {
		resp.Digests = []domain.Digest{}
	}
	if resp.WeeklyHistory == nil {
		resp.WeeklyHistory = []domain.WeekCount{}
	}
	if resp.TopTopics == nil {
		resp.TopTopics = []string{}
	}

	WriteJSON(w, http.StatusOK, resp)
}
