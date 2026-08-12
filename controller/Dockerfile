FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY controller ./controller
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ephemeral-api ./controller/cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ephemeral-worker ./controller/cmd/worker

FROM hashicorp/terraform:1.15.8
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/ephemeral-api /usr/local/bin/ephemeral-api
COPY --from=build /out/ephemeral-worker /usr/local/bin/ephemeral-worker
COPY terraform ./terraform
RUN mkdir -p /app/data /app/terraform/runtime-data
ENTRYPOINT ["/usr/local/bin/ephemeral-api"]
