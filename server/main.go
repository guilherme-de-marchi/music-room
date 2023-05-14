package main

import (
	"log"
)

func main() {
	s := NewServer("tcp", ":8080", "123", 50)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
