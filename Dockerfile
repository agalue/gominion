FROM golang:1.22-alpine3.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

COPY ./ /app/
RUN GOOS=linux go build -o gominion .

FROM alpine:3.19

COPY --from=builder /app/gominion /usr/local/bin/gominion

RUN apk --no-cache add tzdata ca-certificates libcap libcap2 bash && \
    addgroup -S gominion && adduser -S gominion -G gominion && \
    setcap cap_net_raw+ep /usr/local/bin/gominion

USER gominion
LABEL maintainer="Alejandro Galue <agalue@opennms.com>" name="OpenNMS Minion"

ENTRYPOINT [ "/usr/local/bin/gominion" ]
