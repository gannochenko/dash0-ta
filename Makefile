bootstrap:
	@go install github.com/joho/godotenv/cmd/godotenv@latest

run_log_processor:
	cd apps/log-processor && godotenv -f ../../.env.local go run ./cmd/main.go

run:
	docker compose up --build

run_loadtest:
	k6 run ./apps/log-bombarder/index.js

run_loadtest_long:
	k6 run -e DURATION=5m ./apps/log-bombarder/index.js
