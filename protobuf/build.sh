#!/bin/bash

type protoc >/dev/null 2>&1 || { echo >&2 "protoc required but it's not installed; aborting."; exit 1; }

protoc -I . ./proto/sink-message.proto --go_out=./sink
protoc -I . ./proto/kafka-rpc.proto --go_out=./rpc
protoc -I . ./proto/ipc.proto --go_out=plugins=grpc:./ipc
