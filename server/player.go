package main

type music struct {
	name,
	path string
}

type player struct {
	playlist     []music
	currentMusic []byte
}
