package main

import (
	"chess-api/ws"
	"log"
	"net/http"
)

const (
	HOST = "localhost"
	PORT = "3502"
)

func main() {
	m := ws.NewManager()

	router := http.NewServeMux()
	router.HandleFunc("/ws", m.HandleConnection)

	log.Fatalln(http.ListenAndServe(HOST+":"+PORT, router))
}
