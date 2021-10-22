FROM golang:alpine AS builder
RUN  apk add --no-cache alpine-sdk
WORKDIR /app
ADD ./ /app/
RUN echo "@edgecommunity http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache alpine-sdk git cyrus-sasl-dev librdkafka-dev@edgecommunity && \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags static_all,netgo,musl -o gominion .

FROM alpine
COPY --from=builder /app/gominion /usr/local/bin/gominion
RUN apk add --no-cache libcap tzdata bash && \
    addgroup -S minion && \
    adduser -S -G minion minion && \
    setcap cap_net_raw+ep /usr/local/bin/gominion
USER minion
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" name="OpenNMS Minion"
ENTRYPOINT [ "/usr/local/bin/gominion" ]
