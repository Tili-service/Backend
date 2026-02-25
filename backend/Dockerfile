FROM golang:alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o tili-backend cmd/api/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/tili-backend .
EXPOSE 8080
CMD ["./tili-backend"]