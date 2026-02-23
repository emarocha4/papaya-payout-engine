.PHONY: help run test fmt linter migrate-up migrate-down docker-up docker-down

help:
	@echo "Available targets:"
	@echo "  run          - Run the application"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  linter       - Run linter"
	@echo "  docker-up    - Start PostgreSQL"
	@echo "  docker-down  - Stop PostgreSQL"
	@echo "  migrate-up   - Run migrations"
	@echo "  migrate-down - Rollback migrations"

run:
	go run main.go

test:
	go test -v ./...

fmt:
	go fmt ./...

linter:
	golangci-lint run

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	@echo "Applying migrations..."
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000001_create_merchants.up.sql
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000002_create_decisions.up.sql
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000003_create_batches.up.sql
	@echo "Migrations applied successfully"

migrate-down:
	@echo "Rolling back migrations..."
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000003_create_batches.down.sql
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000002_create_decisions.down.sql
	@PGPASSWORD=papaya_pass psql -h localhost -U papaya_user -d papaya_payout_engine -f migration/000001_create_merchants.down.sql
	@echo "Migrations rolled back successfully"
