FROM golang:alpine AS builder
WORKDIR /app
ADD ./ /app/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags musl -a -o gominion

FROM alpine
COPY --from=builder /app/gominion /usr/loca/bin/gominion
RUN apk add --no-cache libcap && \
    addgroup -S minion && \
    adduser -S -G minion minion && \
    setcap cap_net_raw+ep /usr/loca/bin/gominion
USER minion
LABEL maintainer="Alejandro Galue <agalue@opennms.org>" name="OpenNMS Minion"
ENTRYPOINT [ "/usr/loca/bin/gominion" ]
