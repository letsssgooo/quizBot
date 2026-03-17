FROM golang:1.25-bookworm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/quizbot ./cmd/quizbot/main.go

CMD ["/app/quizbot"]