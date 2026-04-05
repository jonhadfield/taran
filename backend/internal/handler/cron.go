package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/digest"
)

type CronHandler struct {
	Scheduler            *digest.Scheduler
	WeeklySummaryRunner  *digest.WeeklySummaryRunner
}

func (h *CronHandler) TriggerDigests(w http.ResponseWriter, r *http.Request) {
	go h.Scheduler.RunNow()
	if h.WeeklySummaryRunner != nil {
		go h.WeeklySummaryRunner.Run()
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
}
