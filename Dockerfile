# 1. Builder stage (Shared build cache)
FROM golang:alpine AS builder
LABEL maintainer="cylonchau"
WORKDIR /apps
COPY ./ /apps
RUN apk add upx bash make && \
    make build module=pantheon-server && \
    make build module=pantheon-controller

# 2. Server target image
FROM alpine AS server
WORKDIR /apps
COPY --from=builder /apps/target/pantheon-server /usr/bin/
RUN chmod +x /usr/bin/pantheon-server
VOLUME ["/apps"]
ENTRYPOINT ["pantheon-server", "--sql-driver=mysql", "--config", "/etc/pantheon/config.toml", "-v", "10"]
EXPOSE 8899/tcp

# 3. Controller target image
FROM alpine AS controller
WORKDIR /apps
COPY --from=builder /apps/target/pantheon-controller /usr/bin/
RUN chmod +x /usr/bin/pantheon-controller
ENTRYPOINT ["pantheon-controller"]