package protection

import (
	"net"
	"time"
)

type SlowRequestProtector struct {
	minReadRate int64
	timeout     time.Duration
}

func NewSlowRequestProtector(minRate int64, timeout time.Duration) *SlowRequestProtector {
	return &SlowRequestProtector{
		minReadRate: minRate,
		timeout:     timeout,
	}
}

type protectedConn struct {
	net.Conn
	protector *SlowRequestProtector
	readBytes int64
	startTime time.Time
}

func (p *SlowRequestProtector) Protect(conn net.Conn) net.Conn {
	return &protectedConn{
		Conn:      conn,
		protector: p,
		startTime: time.Now(),
	}
}

func (c *protectedConn) Read(b []byte) (n int, err error) {
	// Set deadline for read operation
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.protector.timeout)); err != nil {
		return 0, err
	}

	n, err = c.Conn.Read(b)
	c.readBytes += int64(n)

	// Check read rate
	elapsed := time.Since(c.startTime).Seconds()
	if elapsed > 0 {
		rate := float64(c.readBytes) / elapsed
		if rate < float64(c.protector.minReadRate) {
			return n, ErrSlowConnection
		}
	}

	return n, err
}
