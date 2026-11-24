FROM golang:1.27.4-alpine AS build

ARG version=unknown

RUN apk update && \
		apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
		go mod verify

COPY . .

RUN module_path=$(go list -m) && \
	go build \
		-o /go/bin/discord-schema-uploader \
		-ldflags "-X ${module_path}/cmd/version.version=$version" \
		.

FROM alpine:3.24

RUN addgroup -S app && adduser -S -G app app

COPY --from=build /go/bin/discord-schema-uploader /usr/local/bin/discord-schema-uploader

WORKDIR /var/lib/discord-schema-uploader
RUN chown app:app /var/lib/discord-schema-uploader

USER app

ENTRYPOINT ["/usr/local/bin/discord-schema-uploader"]
