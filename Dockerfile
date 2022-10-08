FROM golang:1.16.3 as BUILDER

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitlab-sync-repo
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitlab-sync-repo .
RUN tar -xf ./sync/tools/obsutil.tar.gz

# copy binary config and utils
FROM alpine:3.14
RUN apk update && apk add --no-cache \
        git \
        bash \
        libc6-compat
COPY --from=BUILDER /go/src/github.com/opensourceways/robot-gitlab-sync-repo/robot-gitlab-sync-repo /opt/app/robot-gitlab-sync-repo
COPY --from=BUILDER /go/src/github.com/opensourceways/robot-gitlab-sync-repo/obsutil /opt/app/obsutil
COPY --from=BUILDER /go/src/github.com/opensourceways/robot-gitlab-sync-repo/sync/tools/sync_files.sh /opt/app/sync_file.sh

ENTRYPOINT ["/opt/app/robot-gitlab-sync-repo"]
