# Build
FROM golang:1.18-alpine AS build

RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o /process main.go

# Run
FROM scratch AS run

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /process /process

EXPOSE 8080

ENTRYPOINT ["/process"]
