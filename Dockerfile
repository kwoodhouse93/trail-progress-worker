# Build
FROM golang:1.18-alpine AS build

WORKDIR /build
COPY ./ ./
RUN go mod download
RUN CGO_ENABLED=0 go build -o /process main.go

# Run
FROM scratch

COPY --from=build /process /process

EXPOSE 8080

ENTRYPOINT ["/process"]
