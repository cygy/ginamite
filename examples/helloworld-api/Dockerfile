# Build from the root of the repository:
#   docker build -f examples/helloworld-api/Dockerfile -t ginamite-helloworld-api .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /helloworld-api ./examples/helloworld-api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /helloworld-api /app/helloworld-api
COPY --from=build /src/data/zone1970.tab /app/data/zone1970.tab
COPY examples/helloworld-api/conf /app/conf
EXPOSE 8080
ENTRYPOINT ["/app/helloworld-api", "-api", "-name", "helloworld-api", "-env", "production", "-conf-server", "/app/conf/server.yaml"]
