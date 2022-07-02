.PHONY: up-db
up-db:
	docker-compose build
	docker-compose up -d postgres

.PHONY: down-db
down-db:
	docker-compose down postgres