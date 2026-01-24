FROM golang:1.25.6-alpine3.23 AS backend

ARG VERSION=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY .git .git/
COPY *.go build.sh ./
COPY internal internal/

RUN apk add --no-cache bash curl git && \
  ./build.sh -v "$VERSION"

FROM alpine:3.23.2

LABEL org.opencontainers.image.authors='dev@codehat.de' \
      org.opencontainers.image.url='https://github.com/kodehat/netcupscp-exporter' \
      org.opencontainers.image.documentation='https://github.com/kodehat/netcupscp-exporter' \
      org.opencontainers.image.source='https://github.com/kodehat/netcupscp-exporter' \
      org.opencontainers.image.vendor='kodehat' \
      org.opencontainers.image.licenses='MIT'

WORKDIR /opt

COPY --from=backend /app/netcupscp-exporter ./netcupscp-exporter

# "curl" is added only for Docker healthchecks!
RUN apk add --no-cache tzdata curl && \
  adduser -D -H nonroot && \
  chmod +x ./netcupscp-exporter

EXPOSE 2008/tcp

USER nonroot:nonroot

ENTRYPOINT [ "/opt/netcupscp-exporter" ]