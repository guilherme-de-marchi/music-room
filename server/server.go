package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	ErrNameInUse        = errors.New("username already in use")
	ErrEmptyInput       = errors.New("empty input received")
	ErrInvalidArguments = errors.New("invalid arguments received")
	ErrNoArguments      = errors.New("no arguments received")
	ErrCommandNotFound  = errors.New("command not found")
	ErrFileNotFound     = errors.New("file not found")
)

type clientError struct {
	conn net.Conn
	msg  string
}

func newClientError(c net.Conn, err error) clientError {
	return clientError{
		conn: c,
		msg:  err.Error(),
	}
}

func (e clientError) Error() string {
	return e.msg
}

type server struct {
	network,
	address,
	secret string
	player   player
	commands map[string]command
	clients  map[string]net.Conn
	admins   []string
	bufSize  int
	errChan  chan error
	mu       *sync.Mutex
}

func NewServer(network, address, secret string, bufSize int) *server {
	return &server{
		network: network,
		address: address,
		secret:  secret,
		commands: loadCommands(
			register,
			broadcast,
			list,
			playerCommand,
		),
		clients: make(map[string]net.Conn),
		bufSize: bufSize,
		errChan: make(chan error),
		mu:      new(sync.Mutex),
	}
}

func (s *server) Start() error {
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	go s.handleErrors()

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConnWithError(c)
	}
}

func (s *server) handleErrors() {
	for err := range s.errChan {
		if se, ok := err.(clientError); ok {
			if err = s.send(se.conn, se.msg); err == nil {
				continue
			}
		}
		log.Println(err)
	}
}

func (s *server) handleConn(c net.Conn) error {
	defer c.Close()
	for {
		buf := make([]byte, s.bufSize)
		n, err := c.Read(buf)
		if err != nil {
			return err
		}
		buf = buf[:n]

		in := string(buf)
		in = strings.TrimSuffix(in, "\r\n")
		if in == "" {
			s.errChan <- newClientError(c, ErrEmptyInput)
			continue
		}

		tokens := strings.Split(in, " ")

		cmd, ok := s.commands[tokens[0]]
		if !ok {
			s.errChan <- newClientError(c, ErrCommandNotFound)
			continue
		}

		if err := cmd.f(s, c, tokens); err != nil {
			if _, ok := err.(clientError); ok {
				s.errChan <- err
				continue
			}
			return err
		}
	}
}

func (s *server) handleConnWithError(c net.Conn) {
	if err := s.handleConn(c); err != nil {
		s.errChan <- err
	}
}

func (s *server) addAdmin(name string) error {
	if s.isAdmin(name) {
		return ErrNameInUse
	}

	s.mu.Lock()
	s.admins = append(s.admins, name)
	s.mu.Unlock()

	return nil
}

func (s *server) send(c net.Conn, a ...any) error {
	_, err := c.Write([]byte(fmt.Sprint(a...) + "\r\n"))
	return err
}

func (s *server) subscribe(c net.Conn, name string) error {
	if s.isRegistered(name) {
		return ErrNameInUse
	}

	s.mu.Lock()
	s.clients[name] = c
	s.mu.Unlock()

	return nil
}

func (s *server) broadcast(a ...any) error {
	var err error
	for _, c := range s.clients {
		if e := s.send(c, a...); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func (s *server) isAdmin(name string) bool {
	for _, n := range s.admins {
		if n == name {
			return true
		}
	}
	return false
}

func (s *server) isRegistered(name string) bool {
	_, ok := s.clients[name]
	return ok
}

func (s *server) addMusic(name, path string) {
	s.mu.Lock()
	s.player.playlist = append(s.player.playlist, music{
		name: name,
		path: path,
	})
	s.mu.Unlock()
}

func getKeys[T comparable, U any](m map[T]U) []T {
	l := make([]T, len(m))
	for k := range m {
		l = append(l, k)
	}
	return l
}
