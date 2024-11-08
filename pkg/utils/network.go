package utils

import (
	"io"
	"net"
	"time"
)

// ReadFull reads all bytes with a timeout
func ReadFull(conn net.Conn, buf []byte, timeout time.Duration) error {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	_, err := io.ReadFull(conn, buf)
	return err
}

// WriteFull writes all bytes with a timeout
func WriteFull(conn net.Conn, data []byte, timeout time.Duration) error {
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	_, err := conn.Write(data)
	return err
}
