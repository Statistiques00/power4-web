FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o power4 .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/power4 ./power4
COPY templates ./templates
COPY style.css favicon.svg ./

EXPOSE 8081

CMD ["./power4"]
