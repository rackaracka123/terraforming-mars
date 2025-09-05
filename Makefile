.PHONY: run frontend backend cli test clean

run:
	cd frontend && npm start & cd backend && go run cmd/server/main.go

frontend:
	cd frontend && npm start

backend:
	cd backend && go run cmd/server/main.go

cli:
	cd backend && go build -o bin/tm cmd/cli/main.go && ./bin/tm

cli-install:
	cd backend && go build -o bin/tm cmd/cli/main.go && sudo cp bin/tm /usr/local/bin/tm

test:
	cd backend && go test ./...
	cd frontend && npm test

clean:
	cd backend && go clean
	cd frontend && npm run build