package server

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/timsolov/fragmented-tcp/conf"
	"github.com/timsolov/fragmented-tcp/protocols/lowproto"
)

func TestServer(t *testing.T) {
	config := conf.New()

	server := NewServer(":2000", config.LOG())
	defer server.Stop()

	// CLIENT #1

	conn, err := net.Dial("tcp", ":2000")
	assert.NoError(t, err)

	client1 := lowproto.New(conn)
	defer client1.Close()

	t.Run("Hi required", func(t *testing.T) {
		resp := sendRecv(t, client1, "MSG client2 message")
		assert.Equal(t, "ERROR HI required", resp)
	})

	t.Run("HI client1", func(t *testing.T) {
		resp := sendRecv(t, client1, "HI client1")
		assert.Equal(t, "OK client1", resp)
	})

	// CLIENT #2

	conn, err = net.Dial("tcp", ":2000")
	assert.NoError(t, err)

	client2 := lowproto.New(conn)
	defer client1.Close()

	t.Run("HI client2", func(t *testing.T) {
		resp := sendRecv(t, client2, "HI client2")
		assert.Equal(t, "OK client2", resp)
	})

	t.Run("client1 send message to client2", func(t *testing.T) {
		resp := sendRecv(t, client1, "MSG client2 Hi client1 !!!")
		assert.Equal(t, "OK client2", resp)

		resp = recv(t, client2)
		assert.Equal(t, "MSG client1 Hi client1 !!!", resp)
	})

	t.Run("client2 send message to client1", func(t *testing.T) {
		resp := sendRecv(t, client2, "MSG client1 Hi client1 !!!")
		assert.Equal(t, "OK client1", resp)

		resp = recv(t, client1)
		assert.Equal(t, "MSG client2 Hi client1 !!!", resp)
	})

	// BROADCASTING TO ALL CLIENTS

	t.Run("wait for broadcasting PING message to all clients from server", func(t *testing.T) {
		resp := recv(t, client1)
		assert.Equal(t, "PING", resp)

		resp = recv(t, client2)
		assert.Equal(t, "PING", resp)
	})
}

func sendRecv(t *testing.T, client lowproto.Conn, msg string) string {
	err := client.WritePacket([]byte(msg))
	assert.NoError(t, err)

	resp, err := client.ReadPacket()
	assert.NoError(t, err)

	return string(resp)
}

func recv(t *testing.T, client lowproto.Conn) string {
	resp, err := client.ReadPacket()
	assert.NoError(t, err)

	return string(resp)
}
