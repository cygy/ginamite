# Build from the root of the repository:
#   docker build -f examples/hello/Dockerfile -t ginamite-hello .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /hello ./examples/hello

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /hello /app/hello
COPY --from=build /src/data/zone1970.tab /app/data/zone1970.tab
COPY examples/hello/conf /app/conf
EXPOSE 8080
ENTRYPOINT ["/app/hello", "-api", "-name", "hello", "-env", "production", "-conf-server", "/app/conf/server.yaml"]
