package icmp

import (
	"math"
	"time"
)

// Options options
type Options struct {
	count int
	size  int
	Log   func(string, ...interface{})

	// Timeout (-W) 参数用于设置单个ICMP回显请求等待回应的超时时间，单位通常也是秒。
	// 如果在指定的时间内没有收到回应，该ICMP请求就会被认为是丢失的。
	// 例如：ping -W 2 192.168.1.1
	// 这个命令将对192.168.1.1进行ping操作，每次发送请求后，如果2秒内没有收到回应，则该请求超时。
	Timeout time.Duration // send packet timeout

	wait time.Duration // wait between sending each packet
}

// Option Options function
type Option func(*Options)

func newOptions(opts []Option, extends ...Option) Options {
	opt := Options{
		size:    56,
		count:   4,
		Log:     func(string, ...interface{}) {},
		Timeout: 1 * time.Second,
		wait:    1 * time.Second,
	}
	for _, o := range opts {
		o(&opt)
	}
	for _, o := range extends {
		o(&opt)
	}
	return opt
}

// Size set size
func Size(size int) Option {
	return func(o *Options) {
		o.size = size
	}
}

// Wait seconds between sending each packet.
func Wait(wait time.Duration) Option {
	return func(o *Options) {
		o.wait = wait
	}
}

// Timeout set timeout
func Timeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// Count count
func Count(count int) Option {
	return func(o *Options) {
		if count == 0 {
			count = math.MaxInt64
		}
		o.count = count
	}
}

// Log ..
func Log(f func(string, ...any)) Option {
	return func(o *Options) {
		o.Log = f
	}
}
