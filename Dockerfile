FROM golang:1.24.5-alpine AS builder

RUN apk update

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/main

FROM gcr.io/distroless/static-debian12 AS runner

COPY --from=builder /app/main /

CMD ["/main"]

FROM golang:1.24.5-alpine AS dev

ENV CGO_ENABLED 0
ENV GO111MODULE auto

RUN apk update && \
    apk add --no-cache bash

WORKDIR /app

RUN go install github.com/air-verse/air@latest

CMD ["air", "-c", ".air.toml"]
