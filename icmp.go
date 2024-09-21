package icmp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"net"
	"time"
)

// ICMP icmp config
type ICMP struct {
	opts []Option
	ID   int
}

func (i *ICMP) Echo(ctx context.Context, addr *net.IPAddr, seq int, options Options) (*EchoStat, error) {
	seq = seq % 65536

	conn, err := net.DialIP("ip4:icmp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("dial fail: %v", err)
	}
	defer conn.Close()

	pkt := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   i.ID,
			Seq:  seq,
			Data: bytes.Repeat([]byte("x"), options.size),
		},
	}
	content, err := pkt.Marshal(nil)
	if err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Now().Add(options.Timeout)); err != nil {
		return nil, err
	}

	now := time.Now()

	size, err := conn.Write(content)
	if err != nil || size != len(content) {
		return nil, fmt.Errorf("sendto: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			ipbuf := make([]byte, 20+size) // ping recv back
			_, err := conn.Read(ipbuf)

			if err != nil {
				var opt *net.OpError
				if errors.As(err, &opt) && opt.Timeout() {
					return nil, fmt.Errorf("request timeout for icmp_seq %d", seq)
				}
				return nil, fmt.Errorf("read Fail: %w", err)
			}

			head, err := ipv4.ParseHeader(ipbuf)
			if err != nil {
				return nil, err
			}

			if head.Src.String() != addr.String() {
				continue
			}

			msg, err := icmp.ParseMessage(ICMPv4, ipbuf[head.Len:])
			if err != nil {
				return nil, err
			}

			switch msg.Type {
			case ipv4.ICMPTypeEcho:
				continue
			case ipv4.ICMPTypeEchoReply:
				echo, ok := msg.Body.(*icmp.Echo)
				if !ok {
					return nil, errors.New("ping recv err reply data")
				}
				if echo.ID != msg.Body.(*icmp.Echo).ID || seq != echo.Seq {
					continue
				}
				return &EchoStat{Seq: echo.Seq, TTL: head.TTL, Cost: time.Since(now)}, nil

			case ipv4.ICMPTypeDestinationUnreachable:
				return nil, fmt.Errorf("from %s icmp_seq=%d Destination Unreachable", addr, seq)
			case ipv4.ICMPTypeTimeExceeded:
				return nil, errors.New("TimeExceeded")
			default:
				return nil, fmt.Errorf("not ICMPTypeEchoReply seq=%d, %#v, %#v", seq, msg, msg.Body)
			}
		}
	}
}

// Ping without stdout
func (i *ICMP) Ping(ctx context.Context, host string, opts ...Option) (*Statistics, error) {
	options, stats := newOptions(i.opts, opts...), &Statistics{Host: host, Loss: -1}

	addr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		options.Log("ping: %s", err.Error())
		return stats, err
	}

	options.Log("PING %s (%s) %d(%d) bytes of data.", host, addr, options.size, options.size+28)
	timer := time.NewTicker(1)
	for seq := 0; seq < options.count; seq++ {
		select {
		case <-ctx.Done():
			return stats, ctx.Err()
		case <-timer.C:
			timer.Reset(options.wait)

			stat, err := i.Echo(ctx, addr, seq, options)
			if err != nil {
				options.Log("%v", err)
			} else {
				options.Log("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms", options.size, addr, stat.Seq, stat.TTL, stat.Cost.Seconds()*1000)
			}
			stats.update(stat)
		}
	}

	return stats, err
}
