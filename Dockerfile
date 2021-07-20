FROM golang:alpine AS builder
RUN  apk add --no-cache alpine-sdk
WORKDIR /app
ADD ./ /app/
RUN  GOOS=linux GOARCH=amd64 go build -tags musl -a -o gominion .

FROM alpine
COPY --from=builder /app/gominion /usr/local/bin/gominion
RUN apk add --no-cache libcap tzdata bash && \
    addgroup -S minion && \
    adduser -S -G minion minion && \
    setcap cap_net_raw+ep /usr/local/bin/gominion
USER minion
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" name="OpenNMS Minion"
ENTRYPOINT [ "/usr/local/bin/gominion" ]
