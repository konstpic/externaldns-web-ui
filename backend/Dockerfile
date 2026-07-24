FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /externaldns-web-ui ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /externaldns-web-ui /externaldns-web-ui
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/externaldns-web-ui"]
