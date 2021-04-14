# gominion [![Go Report Card](https://goreportcard.com/badge/github.com/agalue/gominion)](https://goreportcard.com/report/github.com/agalue/gominion)

An implementation of the OpenNMS Minion in Go.

This project started as a proof of concept to understand how hard it would be to reimplement the OpenNMS Minion in Go. Using low-powered devices like a Raspberry Pi as the Minion server could be a possibility. Still, the current Minion is very resource-demanding in typical production environments.

The Java-based one has many features that the GO version is currently missing but hopefully will be added soon.

Kafka must be the broker technology used for the OpenNMS IPC API (both RPC and Sink), with the `single-topic` feature for RPC must be enabled.

For the gRPC server, you could use:

* The one [embedded](https://docs.opennms.org/opennms/releases/27.1.1/guide-install/guide-install.html#_configure_opennms_horizon_2) in OpenNMS.
* The standalone one implemented in [Java](https://github.com/OpenNMS/grpc-server).
* The standalone one implemented in [Go](https://github.com/agalue/onms-grpc-server).

Alternatively, you can use Kafka directly. Although, you'd need Horizon 26 (or Merdian 2020) or newer to use this implementation.

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

> SFlow receiver is enabled, but the parser for Sink API has not been implemented.

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
* XML (`XmlCollector`)

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
brokerType: grpc
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

For Transport TLS:

```yaml
brokerUrl: grpc-server:8990
brokerType: grpc
brokerProperties:
  tls-enabled: "true"
  server-certificate-file: "/etc/server.crt"
```

Or,

```yaml
brokerUrl: grpc-server:8990
brokerType: grpc
brokerProperties:
  tls-enabled: "true"
  server-certificate: +|
  -----BEGIN CERTIFICATE-----
  ...
  -----END CERTIFICATE-----
```

To use Kafka instead of GRPC:

```yaml
brokerUrl: kafka-server:9092
brokerType: kafka
```