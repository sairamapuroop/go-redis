package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"redis-go/internal/commands"
	"redis-go/internal/protocol"
	"strings"
)

type Server struct {
	Address  string
	Commands *commands.Registry
}

func (s *Server) ListenAndServe() error {

	ln, err := net.Listen("tcp", s.Address)

	if err != nil {
		return err
	}
	log.Printf("listening on %s", s.Address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}

}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Fprintf(conn, "+OK\r\n")

	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	defer w.Flush()

	var firstCommandIgnored bool

	for {

		arr, err := protocol.ReadArray(r)

		if err != nil {
			fmt.Fprintf(conn, "-ERR resp parse error: %v\r\n", err)
			return
		}

		if len(arr) == 0 {
			conn.Write([]byte("-ERR empty command\r\n"))
			continue
		}

		cmd := strings.ToUpper(arr[0])
		args := arr[1:]
		resp := s.Commands.Execute(cmd, args)

		// --- Ignore redis-cli's startup probe ---
		if !firstCommandIgnored && cmd == "COMMAND" {
			firstCommandIgnored = true
			// log.Println("Ignoring redis-cli startup probe COMMAND DOCS")
			continue
		}
		firstCommandIgnored = true
		// ---------------------------------------

		if _, err := w.WriteString(resp); err != nil {
			log.Println("write error:", err)
			return
		}

		if err := w.Flush(); err != nil {
			log.Println("flush error:", err)
			return
		}

	}
}
