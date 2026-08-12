FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY examples/demo-service ./examples/demo-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/demo ./examples/demo-service

FROM scratch
COPY --from=build /out/demo /demo
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/demo"]
