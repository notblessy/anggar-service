FROM docker.io/golang:1.23.4 AS builder
WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o /server .

FROM gcr.io/distroless/static
WORKDIR /app
COPY --from=builder /server /server

USER 65532:65532

ENTRYPOINT ["/server", "server"]
