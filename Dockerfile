FROM golang:1.19 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

ADD ./ /app/
RUN apt update && GOOS=linux go build -o gominion .

FROM debian:11-slim

COPY --from=builder /app/gominion /usr/local/bin/gominion

RUN apt update && \
    apt install libcap2-bin bash -y && \
    groupadd gominion && \
    useradd -g gominion -r -s /bin/bash gominion && \
    setcap cap_net_raw+ep /usr/local/bin/gominion

USER gominion
LABEL maintainer="Alejandro Galue <agalue@opennms.com>" name="OpenNMS Minion"

ENTRYPOINT [ "/usr/local/bin/gominion" ]
