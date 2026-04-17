# BUILD STAGE
FROM golang:1.26.2-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o volt

# FINAL STAGE
FROM alpine:3.23.4
WORKDIR /app
COPY --from=builder /app/volt .
RUN chown -R nobody:nogroup ./volt
USER nobody
ENTRYPOINT [ "./volt" ]
