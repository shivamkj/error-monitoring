FROM golang:1.26.2-trixie AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM gcr.io/distroless/base-debian13
COPY --from=builder /server /server
ENTRYPOINT ["/server"]
