package handler

import (
	"net/http"

	"github.com/hadfielj/taran/backend/internal/digest"
)

type CronHandler struct {
	Scheduler *digest.Scheduler
}

func (h *CronHandler) TriggerDigests(w http.ResponseWriter, r *http.Request) {
	go h.Scheduler.RunNow()
	WriteJSON(w, http.StatusOK, map[string]string{"status": "triggered"})
}
