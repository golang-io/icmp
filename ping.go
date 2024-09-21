package icmp

import (
	"os"
)

// proto
const (
	ICMPv4 = 1
	ICMPv6 = 58
	//ipv4Proto = map[string]string{"ip": "ip4:icmp", "udp": "udp4"}
	//ipv6Proto = map[string]string{"ip": "ip6:ipv6-icmp", "udp": "udp6"}
)

// New ping
func New(opts ...Option) *ICMP {
	return &ICMP{
		opts: opts,
		ID:   os.Getpid() & 0xFFFF,
	}
}
