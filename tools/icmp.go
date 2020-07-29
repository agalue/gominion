package tools

import (
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// ProtocolICMP protocol idenification for ICMP IPv4
const ProtocolICMP = 1

// ProtocolIPv6ICMP protocol idenification for ICMP IPv6
const ProtocolIPv6ICMP = 58

// Ping execute an ICMP request against the provided address
// FIXME add logic for handling retries
func Ping(addr string, timeout time.Duration) (time.Duration, error) {
	// Start listening for icmp replies
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, err
	}
	defer c.Close()

	// Resolve any DNS (if used) and get the real IP of the target
	dst, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		return 0, err
	}

	// Make a new ICMP message
	m := icmp.Message{
		Type: ipv4.ICMPTypeEcho, Code: 0,
		Body: &icmp.Echo{
			ID: os.Getpid() & 0xffff, Seq: 1, //<< uint(seq), // TODO
			Data: []byte(""),
		},
	}
	b, err := m.Marshal(nil)
	if err != nil {
		return 0, err
	}

	// Send the request
	start := time.Now()
	n, err := c.WriteTo(b, dst)
	if err != nil {
		return 0, err
	} else if n != len(b) {
		return 0, fmt.Errorf("got %v; want %v", n, len(b))
	}

	// Wait for a reply
	reply := make([]byte, 1500)
	err = c.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return 0, err
	}
	n, peer, err := c.ReadFrom(reply)
	if err != nil {
		return 0, err
	}
	duration := time.Since(start)

	// Parse response
	rm, err := icmp.ParseMessage(ProtocolICMP, reply[:n])
	if err != nil {
		return 0, err
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply:
		return duration, nil
	default:
		return 0, fmt.Errorf("got %+v from %v; want echo reply", rm, peer)
	}
}
