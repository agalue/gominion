FROM golang:alpine AS builder
WORKDIR /app
ADD ./ /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl -a -o gominion

FROM alpine
COPY --from=builder /app/gominion /usr/local/bin/gominion
RUN apk add --no-cache libcap && \
    addgroup -S minion && \
    adduser -S -G minion minion && \
    setcap cap_net_raw+ep /usr/local/bin/gominion
USER minion
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" name="OpenNMS Minion"
ENTRYPOINT [ "/usr/local/bin/gominion" ]
