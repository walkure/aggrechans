FROM --platform=$BUILDPLATFORM golang:1.26.5-alpine3.24 AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app
COPY . /app/
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} go build -o /bin/socket ./socket/
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} go build -o /bin/webhook ./webhook/

FROM scratch AS webhook

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/webhook /webhook
USER 65534:65534

ENV PORT=8080
EXPOSE ${PORT}

ENTRYPOINT ["/webhook"]

# last stage: built when no --target is given
FROM scratch AS socket

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/socket /socket
USER 65534:65534

ENTRYPOINT ["/socket"]
