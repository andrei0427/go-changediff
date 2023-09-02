dev-css: 
	./tailwindcss -i web/views/main.css -o web/static/output.css --watch

dev-widget-css: 
	./tailwindcss -i web/views/widget.css -o web/static/output-widget.css --watch

dev: 
	~/go/bin/air

test:
	go test ./...

run:
	go run cmd/server/main.go

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
	