package main

import (
	"flag"
	"fmt"
	"log"

	"kvstore/server"
	"kvstore/store"
)

func main() {
	port := flag.Int("port", 6379, "port to listen on")
	flag.Parse()

	st := store.New()
	srv := server.New(fmt.Sprintf("localhost:%d", *port), st)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
