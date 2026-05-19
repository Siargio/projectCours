.PHONY: run build test migrate up down

shortener-run:
	@go mod tidy && \
	go run ./cmd/shortener/main.go

shortener-build:
	@go build -o bin/shortener ./cmd/shortener

shortener-test:
	@go test -race -v ./...

shortener-migrate:
	@docker exec -i projectcours-db-1 psql -U postgres -d shortener < migrations/001_create_links.sql

shortener-up:
	@docker compose up -d

shortener-down:
	@docker compose down

shortener-clean:
	@read -p "Очистить все volume файлы в окружения? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		rm -rf bin/; \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi