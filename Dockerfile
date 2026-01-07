FROM golang:alpine AS builder
MAINTAINER cylonchau
WORKDIR /apps
COPY ./ /apps
RUN \
    sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add upx bash make && \
    make build module=pantheon-server && \
    upx -1 _output/pantheon-server && \
    chmod +x _output/pantheon-server

FROM alpine AS runner
WORKDIR /apps
COPY --from=builder /apps/_output/pantheon-server /usr/bin/
RUN chmod +x /usr/bin/pantheon-server
VOLUME ["/apps" ]
ENTRYPOINT ["pantheon-server", "--sql-driver=mysql", "--config", "/etc/pantheon/config.toml", "-v", "10"]
EXPOSE 8899/tcp