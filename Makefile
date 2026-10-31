


deploy-service:
	docker compose up --build --wait
	migrate -path "./internal/database-service/migrations" -database "postgres://app:app_dev_password@localhost:5432/task_service?sslmode=disable" up