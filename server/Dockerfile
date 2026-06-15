FROM golang:1.26-alpine AS build

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /api /app/api

EXPOSE 8080 9090
CMD ["/app/api"]
