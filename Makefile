# Makefile for FinanceBot
# Usage: make [target]
# Targets: up, build, down, logs, clean, env, help

# Tools
COMPOSE := docker compose
ENV_FILE := .env

# Check if .env file exists, create a template if missing
ifeq (,$(wildcard $(ENV_FILE)))
    $(warning File $(ENV_FILE) not found. Creating a template...)
    $(shell echo "TELEGRAM_TOKEN=your_telegram_bot_token\nREDIS_PASSWORD=your_redis_password" > $(ENV_FILE))
endif

# Check if docker-compose is available
CHECK_DOCKER := $(shell which docker-compose || which docker compose)
ifeq (,$(CHECK_DOCKER))
    $(error docker-compose is not installed. Please install Docker Desktop or docker-compose)
endif

# Phony targets
.PHONY: help up build down logs clean env

help:
	@echo "FinanceBot — Development Commands"
	@echo ""
	@echo "  make up        Build and start containers"
	@echo "  make build     Build the bot image only"
	@echo "  make down      Stop containers"
	@echo "  make logs      View bot logs (last 100 lines)"
	@echo "  make clean     Stop and remove containers and perform system cleanup"
	@echo "  make env       Display environment variables from .env"
	@echo ""

up: build
	@echo "Starting bot and Redis..."
	$(COMPOSE) up --remove-orphans

build:
	@echo "Building Docker image for bot..."
	$(COMPOSE) build bot

down:
	@echo "Stopping containers..."
	$(COMPOSE) down

logs:
	@echo "Fetching bot logs..."
	$(COMPOSE) logs --tail=100 bot

clean: down
	@echo "Performing system cleanup..."
	docker system prune -f
	@echo "Cleanup completed"

env:
	@echo "Environment variables from .env:"
	@cat .env 2>/dev/null || echo "Error: .env file not found"