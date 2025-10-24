FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o kvstore ./cmd/kvd/main.go






# ....................

FROM debian:bookworm-slim

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/kvstore .

EXPOSE 6379


ENTRYPOINT [ "./kvstore" ]