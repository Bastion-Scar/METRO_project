FROM golang:1.20-alpine AS build

RUN apk add --no-cache git make ca-certificates && update-ca-certificates

WORKDIR /app


COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app C:\Users\ivan.koptelov\Documents\Database web project\main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/app /usr/local/bin/

WORKDIR /

CMD ["/usr/local/bin/app"]

EXPOSE 8080 