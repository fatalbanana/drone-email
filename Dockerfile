ARG ALPINE_VERSION
ARG BUILD_IMAGE_TAG
ARG GOARCH

FROM golang:${BUILD_IMAGE_TAG} as builder
ENV GOARCH=$GOARCH

WORKDIR /go/src/drone-email
COPY . .

RUN GOOS=linux GOARCH=${GOARCH} CGO_ENABLED=0 go build

ARG  ALPINE_VERSION
FROM alpine:${ALPINE_VERSION}

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /go/src/drone-email/drone-email /bin/
ENTRYPOINT ["/bin/drone-email"]
