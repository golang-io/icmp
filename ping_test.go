package icmp

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test_Ping(t *testing.T) {
	ping := New(Count(100), Log(func(f string, v ...any) {
		fmt.Printf(fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05.000"), f), v...)
	}))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	stats, err := ping.Ping(ctx, "qq.com")
	t.Logf("%s, err=%v", stats, err)
	t.Log(stats.Print())
}
