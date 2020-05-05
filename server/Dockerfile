# syntax=docker/dockerfile:1.0.0-experimental
# Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make g++ musl-dev linux-headers git ca-certificates
RUN --mount=type=secret,id=github_access_token,required \
    git config --global url."https://`cat /run/secrets/github_access_token`:@github.com/".insteadOf "https://github.com/"

WORKDIR /goCeler-oss
ADD . /goCeler-oss
RUN go build -o /tmp/server server/server.go

# Second stage container for deployment
FROM alpine:latest

COPY --from=builder /tmp/server /usr/local/bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["server"]
