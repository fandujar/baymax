FROM golang:1.22.0 as builder

WORKDIR /app
COPY . /app/

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o baymax /app/cmd/baymax/

FROM ubuntu as final
LABEL org.opencontainers.image.source https://github.com/fandujar/baymax
WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /app/baymax /app/
CMD ["./baymax"]