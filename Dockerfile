FROM golang:alpine AS builder
MAINTAINER cylonchau
WORKDIR /apps
COPY ./ /apps
RUN \
    apk add upx bash make && \
    make build module=pantheon-server && \
    upx -1 target/pantheon-server && \
    chmod +x target/pantheon-server

FROM alpine AS runner
WORKDIR /apps
COPY --from=builder /apps/target/pantheon-server /usr/bin/
RUN chmod +x /usr/bin/pantheon-server
VOLUME ["/apps" ]
ENTRYPOINT ["pantheon-server", "--sql-driver=mysql", "--config", "/etc/pantheon/config.toml", "-v", "10"]
EXPOSE 8899/tcp