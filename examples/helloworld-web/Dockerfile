# Build from the root of the repository:
#   docker build -f examples/helloworld-web/Dockerfile -t ginamite-helloworld-web .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /helloworld-web ./examples/helloworld-web

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /helloworld-web /app/helloworld-web
COPY examples/helloworld-web/conf /app/conf
COPY examples/helloworld-web/localization /app/localization
COPY examples/helloworld-web/templates /app/templates
EXPOSE 8080
ENTRYPOINT ["/app/helloworld-web", "-web", "-name", "helloworld-web", "-env", "production", "-conf-server", "/app/conf/server.yaml", "-web-routes", "/app/conf/routes.yaml"]
