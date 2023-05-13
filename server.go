package musicroom

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	ErrNameInUse   = errors.New("username already in use")
	ErrEmptyInput  = errors.New("empty input received")
	ErrNoArguments = errors.New("no arguments received")
)

type server struct {
	network,
	address,
	secret string
	clients map[string]net.Conn
	admins  []string
	mu      *sync.Mutex
}

func NewServer(network, address, secret string) *server {
	return &server{
		network: network,
		address: address,
		secret:  secret,
		clients: make(map[string]net.Conn),
		mu:      new(sync.Mutex),
	}
}

func (s *server) Start() error {
	l, err := net.Listen(s.network, s.address)
	if err != nil {
		return err
	}

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(c)
	}
}

func (s *server) handleConn(c net.Conn) error {
	for {
		buf := make([]byte, 50)
		n, err := c.Read(buf)
		if err != nil {
			return err
		}
		buf = buf[:n]

		in := string(buf)
		in = strings.TrimSuffix(in, "\r\n")
		if in == "" {
			return ErrEmptyInput
		}

		tokens := strings.Split(in, " ")
		lenTokens := len(tokens)

		switch tokens[0] {
		case "register":
			if lenTokens < 2 {
				return ErrNoArguments
			}

			name := tokens[1]
			if err := s.subscribe(c, name); err != nil {
				return err
			}

			if lenTokens > 2 && tokens[2] == s.secret {
				if err := s.addAdmin(name); err != nil {
					return err
				}
			}

		case "broadcast":
			if lenTokens < 2 {
				return ErrNoArguments
			}

			name := tokens[1]
			data, err := os.ReadFile("./musics/" + name)
			if err != nil {
				return err
			}

			log.Println("music " + fmt.Sprint(len(data)))
			if err := s.broadcast([]byte("music " + fmt.Sprint(len(data)))); err != nil {
				return err
			}

			if err := s.broadcast(data); err != nil {
				return err
			}
		}
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

func (s *server) subscribe(c net.Conn, name string) error {
	if _, ok := s.clients[name]; ok {
		return ErrNameInUse
	}

	s.mu.Lock()
	s.clients[name] = c
	s.mu.Unlock()

	return nil
}

func (s *server) broadcast(data []byte) error {
	var err error
	for _, c := range s.clients {
		if _, e := c.Write(data); e != nil {
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
