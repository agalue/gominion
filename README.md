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
* Cisco NX-OS Streaming Telemetry via gRPC
* Netflow5, Netflow9, IPFIX, SFlow
* Graphite

## Detectors

* ICMP (`IcmpDetector`)
* SNMP (`SnmpDetector`)
* TCP (`TcpDetector`)
* HTTP (`HttpDetector`, `HttpsDetector`, `WebDetector`)

## Monitors

* ICMP (`IcmpMonitor`)
* SNMP (`SnmpMonitor`)
* TCP (`TcpMonitor`)
* HTTP (`HttpMonitor`, `HttpsMonitor`, `WebMonitor`)

## Collectors

* HTTP (`HttpCollector`)

> It is important to notice that the SNMP Collector work is handled via the SNMP RPC Module, not by a collector implementation like the rest of them.

## Development

There are skeletons to implement new detectors, monitors and collectors.

Each module folder contains a file called `empty.go` that can be used as a reference.

## Usage

The command configuration can be passed via:

* Environment Variables (using `GOMINION_` as a prefix)
* CLI parameters
* YAML configuration file (defaults to `~/.gominion.yaml`, or can be passed via CLI parameters)

Example YAML configuration:

```yaml
id: go-minion1
location: Apex
brokerUrl: grpc-server:8990
trapPort: 1162
syslogPort: 1514
listeners:
- name: Netflow-5
  port: 8877
  parser: Netflow5UdpParser
- name: Netflow-9
  port: 4729
  parser: Netflow9UdpParser
- name: IPFIX
  port: 4730
  parser: IpfixUdpParser
- name: Graphite
  port: 2003
  parser: ForwardParser
- name: NXOS
  port: 50000
  parser: NxosGrpcParser
- name: SFlow
  port: 6343
  parser: SFlowUdpParser
```

On the above example, `grpc-server` can be a standalone one, or the one embedded with OpenNMS.

> *WARNING:* TLS for gRPC server is not supported.