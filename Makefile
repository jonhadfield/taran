.PHONY: start stop restart \
       start-backend stop-backend restart-backend build-backend \
       start-frontend stop-frontend restart-frontend \
       db stop-db

ROOT_DIR      := $(CURDIR)
BACKEND_PORT  := 8080
FRONTEND_PORT := 3002
PID_DIR       := $(ROOT_DIR)/.pids

# --- All ---

start: start-backend start-frontend

stop: stop-backend stop-frontend

restart: stop start

# --- Backend ---

build-backend:
	@cd $(ROOT_DIR)/backend && go build -o taran ./cmd/taran/
	@echo "Backend built"

start-backend: build-backend
	@mkdir -p $(PID_DIR)
	@if [ -f $(PID_DIR)/backend.pid ] && kill -0 $$(cat $(PID_DIR)/backend.pid) 2>/dev/null; then \
		echo "Backend already running (PID $$(cat $(PID_DIR)/backend.pid))"; \
	else \
		cd $(ROOT_DIR)/backend && TARAN_LOG_SYSLOG=true nohup ./taran > /dev/null 2>&1 & \
		echo $$! > $(PID_DIR)/backend.pid; \
		echo "Backend started (PID $$(cat $(PID_DIR)/backend.pid)), logging to syslog"; \
	fi

stop-backend:
	@if [ -f $(PID_DIR)/backend.pid ]; then \
		pid=$$(cat $(PID_DIR)/backend.pid); \
		if kill -0 $$pid 2>/dev/null; then \
			kill $$pid && echo "Stopped backend (PID $$pid)"; \
		else \
			echo "Backend not running (stale PID $$pid)"; \
		fi; \
		rm -f $(PID_DIR)/backend.pid; \
	else \
		echo "Backend not running (no PID file)"; \
	fi

restart-backend: stop-backend
	@sleep 1
	@$(MAKE) start-backend

# --- Frontend ---

start-frontend:
	@mkdir -p $(PID_DIR)
	@if [ -f $(PID_DIR)/frontend.pid ] && kill -0 $$(cat $(PID_DIR)/frontend.pid) 2>/dev/null; then \
		echo "Frontend already running (PID $$(cat $(PID_DIR)/frontend.pid))"; \
	else \
		cd $(ROOT_DIR)/frontend && nohup npm run dev > /dev/null 2>&1 & \
		echo $$! > $(PID_DIR)/frontend.pid; \
		echo "Frontend started (PID $$(cat $(PID_DIR)/frontend.pid))"; \
	fi

stop-frontend:
	@if [ -f $(PID_DIR)/frontend.pid ]; then \
		pid=$$(cat $(PID_DIR)/frontend.pid); \
		if kill -0 $$pid 2>/dev/null; then \
			kill $$pid && echo "Stopped frontend (PID $$pid)"; \
		else \
			echo "Frontend not running (stale PID $$pid)"; \
		fi; \
		rm -f $(PID_DIR)/frontend.pid; \
	else \
		echo "Frontend not running (no PID file)"; \
	fi

restart-frontend: stop-frontend
	@sleep 1
	@$(MAKE) start-frontend

# --- Database ---

db:
	cd $(ROOT_DIR)/backend && docker compose up -d

stop-db:
	cd $(ROOT_DIR)/backend && docker compose down
