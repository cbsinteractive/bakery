FROM golang:1.13-alpine3.10 AS build_base

ARG SSH_PRIVATE_KEY

RUN apk add --no-cache ca-certificates curl git openssh build-base

RUN mkdir -p ~/.ssh && umask 0077 && echo "${SSH_PRIVATE_KEY}" > ~/.ssh/id_rsa \
	&& git config --global url."git@github.com:".insteadOf https://github.com/ \
	&& ssh-keyscan github.com >> ~/.ssh/known_hosts

RUN go env -w "GONOSUMDB=github.com/cbsinteractive"

WORKDIR /bakery

COPY go.mod .
COPY go.sum .

RUN rm ~/.ssh/id_rsa
RUN rm ~/.ssh/known_hosts

FROM build_base AS builder

COPY . .
RUN go build -mod=vendor -o bakery cmd/http/main.go

FROM alpine:latest

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app/

RUN apk add --no-cache tzdata

COPY --from=builder bakery .

RUN adduser -D bakery
USER bakery

EXPOSE 8080

ENTRYPOINT ["./bakery"]