#!/bin/bash

type protoc >/dev/null 2>&1 || { echo >&2 "protoc required but it's not installed; aborting."; exit 1; }

protoc --proto_path=./proto --go_out=./ sink-message.proto
protoc --proto_path=./proto --go_out=./ kafka-rpc.proto
protoc --proto_path=./proto --go_out=./ telemetry.proto
protoc --proto_path=./proto --go_out=./ netflow.proto
protoc --proto_path=./proto --go_out=./ --go-grpc_out=./ ipc.proto
protoc --proto_path=./proto --go_out=./ --go-grpc_out=./ mdt_dialout.proto
