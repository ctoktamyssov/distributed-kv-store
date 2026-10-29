package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"kvstore/store"
)

type Server struct {
	addr  string
	store *store.Store
}

func New(addr string, st *store.Store) *Server {
	return &Server{addr: addr, store: st}
}

// ListenAndServe accepts connections forever.
func (s *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("listening on %s", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		// TODO: what happens if you call handleConn without `go`?
		// Try it with two netcat clients before adding `go`.
		go s.handleConn(conn)
	}
}

// handleConn reads one command per line and writes one reply per line.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	log.Printf("client connected: %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		reply := s.execute(line)
		fmt.Fprintln(conn, reply)
	}
	log.Printf("client disconnected: %s", conn.RemoteAddr())
}

// execute parses one line and returns the reply string.
func (s *Server) execute(line string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "ERR empty command"
	}

	cmd := strings.ToUpper(parts[0])
	switch cmd {
	case "SET":
		if len(parts) != 3 {
			return "ERR wrong number of args"
		}

		key, value := parts[1], parts[2]
		s.store.Set(key, value)
		return "OK"
	case "GET":
		if len(parts) != 2 {
			return "ERR wrong number of args"
		}

		key := parts[1]
		value, ok := s.store.Get(key)

		if !ok {
			return "NOT_FOUND"
		}
		return value
	case "DEL":
		if len(parts) != 2 {
			return "ERR wrong number of args"
		}

		key := parts[1]
		s.store.Delete(key)
		return "OK"
	default:
		return "ERR unknown command"
	}
}
