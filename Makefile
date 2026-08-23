.PHONY: test api web tidy build-web

test:
	cd backend && go test ./...

tidy:
	cd backend && go mod tidy

api:
	cd backend && go run ./cmd/api

web:
	cd frontend && npm run dev

build-web:
	cd frontend && npm run build

