package main

import (
	"errors"
	"net"
	"os"
)

type commandFunc func(*server, net.Conn, []string) error

type command struct {
	key,
	description string
	f commandFunc
}

var (
	register = command{
		key:         "register",
		description: "",
		f:           registerFunc,
	}

	broadcast = command{
		key:         "broadcast",
		description: "",
		f:           broadcastFunc,
	}
)

func registerFunc(s *server, c net.Conn, tokens []string) error {
	if len(tokens) < 2 {
		return newClientError(c, ErrNoArguments)
	}

	name := tokens[1]
	if err := s.subscribe(c, name); err != nil {
		return newClientError(c, err)
	}

	if len(tokens) > 2 && tokens[2] == s.secret {
		if err := s.addAdmin(name); err != nil {
			return newClientError(c, err)
		}
	}

	return nil
}

func broadcastFunc(s *server, c net.Conn, tokens []string) error {
	if len(tokens) < 2 {
		return newClientError(c, ErrNoArguments)
	}

	name := tokens[1]
	data, err := os.ReadFile("./musics/" + name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newClientError(c, ErrFileNotFound)
		}
		return err
	}

	if err := s.broadcast("music ", len(data)); err != nil {
		return newClientError(c, err)
	}

	if err := s.broadcast(data); err != nil {
		return newClientError(c, err)
	}

	return nil
}

func loadCommands(cmds ...command) map[string]command {
	m := make(map[string]command)

	for _, c := range cmds {
		m[c.key] = c
	}
	return m
}
