# run following from goCeler-oss folder:
# COMMIT=$(git log -1 --pretty=%h) && SERVER=cmnty/server && SERVER_COMMIT=$SERVER:$COMMIT && env DOCKER_BUILDKIT=1 docker build -f server/Dockerfile -t $SERVER_COMMIT -t $SERVER:latest .

# syntax=docker/dockerfile:1.0.0-experimental
# Go builder container
FROM golang:1.12.4-alpine as builder

RUN apk add --no-cache make g++ musl-dev linux-headers git ca-certificates

WORKDIR /goCeler-oss
ADD . /goCeler-oss
RUN go build -o /tmp/server server/server.go && go build -o /tmp/first-time-run-setup setup/main.go && go build -o /tmp/webproxy webproxy/cmd/main.go

# Second stage container for deployment
FROM alpine:latest

COPY --from=builder /tmp/server /usr/local/bin
COPY --from=builder /tmp/webproxy /usr/local/bin
COPY --from=builder /tmp/first-time-run-setup /usr/local/bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["server"]
