# gominion [![Go Report Card](https://goreportcard.com/badge/github.com/agalue/gominion)](https://goreportcard.com/report/github.com/agalue/gominion)

An implementation of the OpenNMS Minion in Go using gRPC

This project started as a proof of concept to understand how hard it would be to reimplement the OpenNMS Minion in Go.

The Java-based one has lots of features that is currently missing, but hopefully will be added soon.

## RPC Modules

* Echo
* DNS
* SNMP
* Ping
* Detect
* Collect
* Poller

## Sink Modules

* Heartbeat
* SNMP Traps (SNMPv1 and SNMPv2)
* Syslog (TCP and UDP)
* NX-OS Streaming Telemetry via gRPC
* Netflow5, Netflow9, IPFIX
* Graphite

## Detectors

* ICMP (`IcmpDetector`)
* SNMP (`SnmpDetector`)
* TCP (`TcpDetector`)
* HTTP (`HttpDetector`, `HttpsDetector`, `WebDetector`)

## Monitors

* ICMP (`IcmpMonitor`)
* SNMP (`SnmpMonitor`)
* TCP (`TcpMonitor`, basic functionality only)
* HTTP (`HttpMonitor`, `HttpsMonitor`, `WebMonitor`)

## Collectors

* HTTP (`HttpCollector`)

It is important to notice that the SNMP Collector work is handled via the SNMP RPC Module, not by a collector implementation like the rest of them.

## Development

There are skeletons to implement new detectors, monitors and collectors.

Each module folder contains a file called `empty.go` that can be used as a reference.
