FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /tech-stack-lookup .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /tech-stack-lookup /usr/bin/tech-stack-lookup
ENTRYPOINT ["/usr/bin/tech-stack-lookup"]
