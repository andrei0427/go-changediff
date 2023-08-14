dev-css: 
	./tailwindcss -i views/main.css -o static/output.css --watch

dev: 
	~/go/bin/air

run:
	go run main.go

migrate:
	go run cmd/db/migrate.go -up

migrate-down:
	go run cmd/db/migrate.go -down

migrate-ver:
	go run cmd/db/migrate.go -v $(ver)

build-css: 
	./tailwindcss -i views/main.css -o static/output.css --minify
	
sql-gen:
	sqlc generate
	