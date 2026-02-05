COMPOSE = docker compose
PROJECT = app

.PHONY: up down rebuild clean logs

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

rebuild:
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f

clean:
	$(COMPOSE) down -v --rmi local --remove-orphans