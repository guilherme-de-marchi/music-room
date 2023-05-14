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
	list = command{
		key:         "list",
		description: "",
		f:           listFunc,
	}

	register = command{
		key:         "register",
		description: "",
		f:           registerFunc,
	}

	playerCommand = command{
		key:         "player",
		description: "",
		f:           playerFunc,
	}

	broadcast = command{
		key:         "broadcast",
		description: "",
		f:           broadcastFunc,
	}
)

func listFunc(s *server, c net.Conn, tokens []string) error {
	if len(tokens) < 2 {
		return newClientError(c, ErrNoArguments)
	}

	switch tokens[1] {
	case "clients":
		return listClientsFunc(s, c, tokens)
	case "admins":
		return listAdminsFunc(s, c, tokens)
	case "playlist":
		return listPlaylistFunc(s, c, tokens)
	case "musics":
		return listMusicsFunc(s, c, tokens)
	default:
		return newClientError(c, ErrInvalidArguments)
	}
}

func listClientsFunc(s *server, c net.Conn, tokens []string) error {
	return s.send(c, getKeys(s.clients))
}

func listAdminsFunc(s *server, c net.Conn, tokens []string) error {
	return s.send(c, s.admins)
}

func listPlaylistFunc(s *server, c net.Conn, tokens []string) error {
	return s.send(c, s.player.playlist)
}

func listMusicsFunc(s *server, c net.Conn, tokens []string) error {
	var musics []string
	entries, err := os.ReadDir("./musics")
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		musics = append(musics, e.Name())
	}

	return s.send(c, musics)
}

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

func playerFunc(s *server, c net.Conn, tokens []string) error {
	if len(tokens) < 2 {
		return newClientError(c, ErrNoArguments)
	}

	switch tokens[1] {
	case "add":
		return playerAddFunc(s, c, tokens)
	default:
		return newClientError(c, ErrInvalidArguments)
	}
}

func playerAddFunc(s *server, c net.Conn, tokens []string) error {
	if len(tokens) < 3 {
		return newClientError(c, ErrNoArguments)
	}

	name := tokens[2]
	entries, err := os.ReadDir("./musics")
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == name {
			s.addMusic(name, "./musics/"+name)
			return nil
		}
	}

	return newClientError(c, ErrFileNotFound)
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

	if err := s.broadcast("music ", name); err != nil {
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
