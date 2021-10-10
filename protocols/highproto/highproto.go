package highproto

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
)

type MessageKind byte

const (
	UNKNOWN MessageKind = iota
	HI
	CLIENTS
	MSG
	PONG
)

// String implementation of Stringer interface
func (k MessageKind) String() string {
	switch k {
	case HI:
		return "HI"
	case CLIENTS:
		return "CLIENTS"
	case MSG:
		return "MSG"
	}
	return "UNKNOWN"
}

// Delimiter is a byte which separates message's octets
const Delimiter = 0x20 // Space

// SYSTEM reserved name for server's name
const SYSTEM = "SYSTEM"

var (
	ErrUnknownPacket = errors.New("unknown packet")
)

// Parse parses byte packet and returns kind of message and parameters.
func Parse(packet []byte) (kind MessageKind, params []string, err error) {
	parts := bytes.SplitN(packet, []byte{Delimiter}, 2) // SplitN to split fine should take minimum 2 as amount of parts
	if len(parts) < 1 {
		return UNKNOWN, nil, errors.Wrap(ErrUnknownPacket, "split first octect")
	}

	octetsAmount := 1
	switch string(parts[0]) {
	case "HI":
		kind = HI
		octetsAmount = 2 // HI <NAME>
	case "CLIENTS":
		kind = CLIENTS
		octetsAmount = 1 // CLIENTS
	case "MSG":
		kind = MSG
		octetsAmount = 3 // MSG <FROM> <TEXT>
	case "PONG":
		kind = PONG
		octetsAmount = 1 // PONG
	default:
		return UNKNOWN, nil, ErrUnknownPacket
	}

	if octetsAmount > 1 {
		parts = bytes.SplitN(packet, []byte{Delimiter}, octetsAmount)
		if len(parts) != octetsAmount {
			return UNKNOWN, nil, errors.Wrap(ErrUnknownPacket, "split whole message")
		}

		params = make([]string, len(parts)-1)
		for i := 1; i < len(parts); i++ {
			params[i-1] = string(parts[i])
		}
	}

	return
}

type ResponseKind byte

const (
	OK ResponseKind = iota + 1
	ERROR
)

// Response builds OK or ERROR response message.
func Response(kind ResponseKind, param string) []byte {
	switch kind {
	case OK:
		return []byte(fmt.Sprintf("OK %s", param))
	case ERROR:
		return []byte(fmt.Sprintf("ERROR %s", param))
	}
	return nil
}

// Msg builds MSG message.
func Msg(from, text string) []byte {
	var b bytes.Buffer
	b.WriteString("MSG")
	b.WriteByte(Delimiter)
	b.WriteString(from)
	b.WriteByte(Delimiter)
	b.WriteString(text)
	return b.Bytes()
}
