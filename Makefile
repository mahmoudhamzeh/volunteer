.PHONY: test api web tidy postman openapi

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

# Import postman/Mahak-Volunteer-Management.postman_collection.json in Postman
postman:
	@echo postman/Mahak-Volunteer-Management.postman_collection.json

openapi:
	python3 scripts/gen_openapi.py

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

# Import postman/Mahak-Volunteer-Management.postman_collection.json in Postman
postman:
	@echo postman/Mahak-Volunteer-Management.postman_collection.json
