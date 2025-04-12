DB_URL=postgresql://admin:secret@localhost:5432/personal_blog_db?sslmode=disable
sqlc:
	sqlc generate

server:
	go run main.go

migrate_up:
	migrate -path internal/db/migrations -database "$(DB_URL)" -verbose up

migrate_down:
	migrate -path internal/db/migrations  -database "$(DB_URL)" -verbose down

server:
	go run cmd/server/main.go

.PHONY: sqlc server migrate_up migrate_down server


