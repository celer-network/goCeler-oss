# syntax=docker/dockerfile:1.0.0-experimental
# Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git ca-certificates
RUN --mount=type=secret,id=github_access_token,required \
    git config --global url."https://`cat /run/secrets/github_access_token`:@github.com/".insteadOf "https://github.com/"

WORKDIR /goCeler-oss
ADD . /goCeler-oss
RUN go build -o /tmp/osp-cli tools/osp-cli/osp_cli.go && go build -o /tmp/channel-view-server tools/view-server/view_server.go

# Second stage container for deployment
FROM alpine:latest

COPY --from=builder /tmp/osp-cli /usr/local/bin
COPY --from=builder /tmp/channel-view-server /usr/local/bin
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /goCeler/tools/view-server/static/* /etc/cv_static/
ENTRYPOINT ["channel-view-server"]
