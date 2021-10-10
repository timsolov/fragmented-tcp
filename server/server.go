package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/timsolov/fragmented-tcp/protocols/highproto"
	"github.com/timsolov/fragmented-tcp/protocols/lowproto"
)

var Version string
var Buildtime string

// Server describes tcp listener with gracefull shutdown
// it was inspired by this article: https://eli.thegreenplace.net/2020/graceful-shutdown-of-a-tcp-server-in-go/
type Server struct {
	listener net.Listener
	log      *logrus.Entry
	quit     chan interface{}
	wg       sync.WaitGroup

	clientNames       map[lowproto.Conn]string // map of client's names (map[socket]name)
	clientConns       map[string]lowproto.Conn // map to prevent duplication of names and to fast request socket by name
	mu                sync.RWMutex
	keepAliveInterval time.Duration
}

// NewServer creates new Server instance
func NewServer(addr string, log *logrus.Entry) *Server {
	s := &Server{
		quit:              make(chan interface{}),
		log:               log,
		clientNames:       make(map[lowproto.Conn]string),
		clientConns:       make(map[string]lowproto.Conn),
		keepAliveInterval: time.Second * 1,
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.log.WithError(err).Fatalf("listen tcp server on %s", addr)
	}
	s.listener = l
	s.wg.Add(2)
	go s.serve()
	go s.keepAlive()
	return s
}

// Stop method to gracefull shutdown tcp listener.
func (s *Server) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) serve() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				s.log.WithError(err).Error("accept error")
			}
		} else {
			s.wg.Add(1)
			go func() {
				c := lowproto.New(conn)
				s.handleConnection(c)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) handleConnection(conn lowproto.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clientConns, s.clientNames[conn])
		delete(s.clientNames, conn)
		s.mu.Unlock()

		conn.Close()
	}()

ReadLoop:
	for {
		select {
		case <-s.quit:
			return
		default:
			packet, err := conn.ReadPacket()
			if err != nil {
				switch errors.Cause(err) {
				case lowproto.ErrTimeout:
					continue ReadLoop
				default:
					s.log.WithError(err).Error("lowproto reading")
				}
			}

			err = s.dispatch(conn, packet)
			if err != nil {
				s.log.WithError(err).Error("dispatch message")
				return
			}
		}
	}
}

func (s *Server) keepAlive() {
	defer s.wg.Done()

	for {
		select {
		case <-s.quit:
			return
		case <-time.After(s.keepAliveInterval): // once a minute
			s.mu.RLock()
			for _, conn := range s.clientConns {
				go func(conn lowproto.Conn) {
					conn.WritePacket(
						[]byte("PING"),
					)
				}(conn)
			}
			s.mu.RUnlock()
		}
	}
}

func (s *Server) dispatch(conn lowproto.Conn, packet []byte) error {
	kind, params, err := highproto.Parse(packet)
	if err != nil {
		return errors.Wrap(err, "parse message")
	}

	var (
		fromName string
		ok       bool
	)

	s.mu.RLock()
	if fromName, ok = s.clientNames[conn]; !ok && kind != highproto.HI {
		s.mu.RUnlock()
		if err = conn.WritePacket(
			highproto.Response(highproto.ERROR, "HI required"),
		); err != nil {
			return fmt.Errorf("writePacket: HI required")
		}
		return nil
	}
	s.mu.RUnlock()

	switch kind {
	case highproto.HI:
		fromName = params[0]
		if fromName == highproto.SYSTEM {
			if err = conn.WritePacket(
				highproto.Response(highproto.ERROR, "not possible to take SYSTEM name"),
			); err != nil {
				return fmt.Errorf("writePacket: not possible to take SYSTEM name")
			}

			return nil
		}

		// prevent duplications
		s.mu.RLock()
		if _, ok = s.clientConns[fromName]; ok {
			s.mu.RUnlock()
			if err = conn.WritePacket(
				highproto.Response(highproto.ERROR, "the name already taken"),
			); err != nil {
				return fmt.Errorf("writePacket: the name already taken")
			}
			return nil
		}
		s.mu.RUnlock()

		// register user
		s.mu.Lock()
		s.clientNames[conn] = fromName
		s.clientConns[fromName] = conn
		s.mu.Unlock()

		if err = conn.WritePacket(
			highproto.Response(highproto.OK, fromName),
		); err != nil {
			return fmt.Errorf("writePacket: OK %s", fromName)
		}

	case highproto.CLIENTS:
		s.mu.RLock()
		names := make([]string, 0, len(s.clientNames))
		for _, name := range s.clientNames {
			names = append(names, name)
		}
		s.mu.RUnlock()

		namesParam := strings.Join(names, "\n")

		if err = conn.WritePacket(
			highproto.Response(highproto.OK, namesParam),
		); err != nil {
			return fmt.Errorf("writePacket: OK %s", namesParam)
		}

	case highproto.MSG:
		var (
			to     lowproto.Conn
			toName string = params[0]
		)

		s.mu.RLock()
		if to, ok = s.clientConns[toName]; !ok {
			s.mu.RUnlock()

			if err = conn.WritePacket(
				highproto.Response(highproto.ERROR, "unknown receiver of message"),
			); err != nil {
				return fmt.Errorf("writePacket: ERROR unknown receiver of message")
			}
			return nil
		}
		s.mu.RUnlock()

		// send to receiver the message
		if err = to.WritePacket(
			highproto.Msg(fromName, params[1]),
		); err != nil {
			s.log.WithError(err).Error("send message to receiver")
			return nil
		}

		// send response to sender
		if err = conn.WritePacket(
			highproto.Response(highproto.OK, toName),
		); err != nil {
			return fmt.Errorf("writePacket: OK %s", toName)
		}

	case highproto.PONG: //skip
		return nil
	}

	return nil
}
