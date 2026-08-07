# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/immichframe-opencloud ./cmd/immichframe-opencloud

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/immichframe-opencloud /usr/local/bin/immichframe-opencloud
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/immichframe-opencloud"]
