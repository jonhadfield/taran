.PHONY: start stop restart \
       start-backend stop-backend restart-backend \
       start-frontend stop-frontend restart-frontend \
       db stop-db

BACKEND_PORT  := 8080
FRONTEND_PORT := 3002

# --- All ---

start: start-backend start-frontend

stop: stop-backend stop-frontend

restart: stop start

# --- Backend ---

start-backend:
	@cd backend && go run ./cmd/taran &
	@echo "Backend starting on port $(BACKEND_PORT)"

stop-backend:
	@pid=$$(lsof -ti tcp:$(BACKEND_PORT) -sTCP:LISTEN); \
	if [ -n "$$pid" ]; then \
		kill $$pid 2>/dev/null && echo "Stopped backend (port $(BACKEND_PORT), PID $$pid)"; \
	else \
		echo "Backend not running on port $(BACKEND_PORT)"; \
	fi

restart-backend: stop-backend
	@sleep 1
	@$(MAKE) start-backend

# --- Frontend ---

start-frontend:
	@cd frontend && npm run dev &
	@echo "Frontend starting on port $(FRONTEND_PORT)"

stop-frontend:
	@pid=$$(lsof -ti tcp:$(FRONTEND_PORT) -sTCP:LISTEN); \
	if [ -n "$$pid" ]; then \
		kill $$pid 2>/dev/null && echo "Stopped frontend (port $(FRONTEND_PORT), PID $$pid)"; \
	else \
		echo "Frontend not running on port $(FRONTEND_PORT)"; \
	fi

restart-frontend: stop-frontend
	@sleep 1
	@$(MAKE) start-frontend

# --- Database ---

db:
	cd backend && docker compose up -d

stop-db:
	cd backend && docker compose down
