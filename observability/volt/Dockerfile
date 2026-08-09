# BUILD STAGE
# https://hub.docker.com/_/golang
FROM golang:1.26.5-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o volt

# FINAL STAGE
# https://hub.docker.com/_/alpine
FROM alpine:3.24.1
WORKDIR /app
COPY --from=builder /app/volt .
RUN chown -R nobody:nogroup ./volt
USER nobody
ENTRYPOINT [ "./volt" ]
