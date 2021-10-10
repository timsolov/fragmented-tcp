package lowproto

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/pkg/errors"
)

// Predefined errors
var (
	ErrTimeout   = errors.New("timeout")
	ErrBadPacket = errors.New("bad packet")
	ErrMismatch  = errors.New("mismatch")
	ErrEOF       = errors.New("EOF")
)

//go:generate mockgen -mock_names=Conn=MockNetConn -destination=conn_mock_test.go -package=lowproto net Conn

// Config for create new Conn
type Config struct {
	ReadLendthTimeout time.Duration
	ReadPacketTimeout time.Duration
}

// Conn main wrapper for net connection
type Conn struct {
	config Config
	conn   net.Conn
}

// option pattern to configure Conn

// ConnOpt option func
type ConnOpt func(c *Conn)

// ReadLengthTimeout set read timeout for length of packet
func ReadLengthTimeout(t time.Duration) ConnOpt {
	return func(c *Conn) {
		c.config.ReadLendthTimeout = t
	}
}

// ReadPacketTimeout set read timeout for body of packet
func ReadPacketTimeout(t time.Duration) ConnOpt {
	return func(c *Conn) {
		c.config.ReadPacketTimeout = t
	}
}

// New creates new net.Conn wrapper to work with fragmented tcp packets
func New(conn net.Conn, opts ...ConnOpt) Conn {
	config := Config{
		ReadLendthTimeout: time.Second * 2,
		ReadPacketTimeout: time.Second * 2,
	}

	c := Conn{
		config: config,
		conn:   conn,
	}

	for _, opt := range opts {
		opt(&c)
	}

	return c
}

// Close implemetation of Closer interface
func (c *Conn) Close() error {
	c.conn.Close()
	return nil
}

// ReadPacket read fragmented packet from underlaying connection.
func (c *Conn) ReadPacket() (packet []byte, err error) {
	bufLength := make([]byte, 2)

	c.conn.SetDeadline(time.Now().Add(c.config.ReadLendthTimeout))

	n, err := c.conn.Read(bufLength)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return nil, ErrTimeout
		} else if err != io.EOF {
			return nil, errors.Wrap(err, "read length bytes")
		}
		return nil, ErrEOF
	}
	if n != 2 {
		return nil, errors.Wrap(ErrBadPacket, "length is not 2 bytes")
	}

	var length uint16

	err = binary.Read(bytes.NewReader(bufLength), binary.BigEndian, &length)
	if err != nil {
		return nil, errors.Wrap(err, "read length from 2 bytes to int16")
	}

	buf := make([]byte, length)
	c.conn.SetDeadline(time.Now().Add(c.config.ReadPacketTimeout))
	n, err = c.conn.Read(buf)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			return nil, ErrTimeout
		} else if err != io.EOF {
			return nil, errors.Wrap(err, "error occurred while reading packet")
		}
		return nil, ErrEOF
	}
	if n != int(length) {
		return nil, errors.Wrap(ErrBadPacket, "broken packet received")
	}

	return buf, nil
}

// WritePacket write fragmented packet to underlaying connection.
func (c *Conn) WritePacket(packet []byte) (err error) {
	lenBuf := make([]byte, 2)
	length := uint16(len(packet))

	binary.BigEndian.PutUint16(lenBuf, length)

	n, err := c.conn.Write(append(lenBuf, packet...))
	if err != nil {
		return errors.Wrap(err, "write to connection")
	}

	if n != 2+len(packet) {
		return errors.Wrap(ErrMismatch, "not all bytes sended")
	}

	return nil
}
