FROM golang:alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY . .
RUN swag init --generalInfo cmd/api/main.go --output docs
RUN go build -o tili-backend cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/tili-backend .
EXPOSE 8000
CMD ["./tili-backend"]