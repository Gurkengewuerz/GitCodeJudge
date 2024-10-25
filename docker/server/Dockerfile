FROM golang:1.23-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /dummy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/main.go

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage
SHELL ["/busybox/sh", "-c"]

WORKDIR /

COPY --from=build-stage /server /server
COPY --from=build-stage --chown=nonroot:nonroot /dummy /db
COPY --from=build-stage --chown=nonroot:nonroot /dummy /test_cases

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/server"]
