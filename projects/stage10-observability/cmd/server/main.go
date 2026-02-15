package main

import (
	"log"
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage10-observability/internal/server"
)

func main() {
	s := server.NewServer()
	log.Fatal(http.ListenAndServe(":18100", s.Handler()))
}
