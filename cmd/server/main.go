package main

import (
	"log"

	musicroom "github.com/Guilherme-De-Marchi/music-room"
)

func main() {
	s := musicroom.NewServer("tcp", ":8080", "123")
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
