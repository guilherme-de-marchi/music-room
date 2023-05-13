package main

import (
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		log.Fatal("not enought args")
	}

	name := args[1]

	c, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = c.Write([]byte("register " + name)); err != nil {
		log.Fatal(err)
	}

	for {
		buf := make([]byte, 50)
		n, err := c.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		data := buf[:n]

		in := string(data)
		if in == "" {
			log.Fatal("void msg")
		}

		tokens := strings.Split(in, " ")
		lenTokens := len(tokens)

		log.Println("received: ", tokens)

		switch tokens[0] {
		case "music":
			if lenTokens < 2 {
				log.Fatal("not enought")
			}

			// size, err := strconv.Atoi(tokens[1])
			// if err != nil {
			// 	log.Fatal(err)
			// }

			dst, err := os.Create("music_a.mp3")
			if err != nil {
				log.Fatal(err)
			}

			n, err := io.Copy(dst, c)
			if err != nil {
				log.Fatal(err)
			}

			// buf = make([]byte, size)
			// n, err = c.Read(buf)
			// if err != nil {
			// 	log.Fatal(err)
			// }

			// data := buf[:n]
			log.Println("music: ", n)
		}
	}

}
