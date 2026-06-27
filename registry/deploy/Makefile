SHELL := /bin/bash

.PHONY: build up down logs migrate

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate:
	docker compose exec registry kylix-registry -migrate