bootstrap:
	@go install github.com/joho/godotenv/cmd/godotenv@latest

run_log_processor:
	cd apps/log-processor && godotenv -f ../../.env.local go run ./cmd/main.go

run:
	docker compose up
