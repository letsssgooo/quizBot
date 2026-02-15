include .env
export

.PHONY: run build test clean migrate-up migrate-down

run:
	go run ${MAIN}

build:
	go build -o bin/${APP_NAME} ${MAIN}

test:
	go test ./...

clean:
	rm -rf bin

migrate-up:
	migrate -path migrations -database ${DB_CONN_URL} up

migrate-down:
	migrate -path migrations -database ${DB_CONN_URL} down