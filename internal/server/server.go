package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"redis-go/internal/commands"
	"redis-go/internal/helper"
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

		cmd, args, ttl, err := helper.ParseCommand(arr)

		log.Println(cmd, args, ttl)

		if err != nil {
			fmt.Fprintf(conn, "-ERR%v\r\n", err)
			w.Flush()
			continue
		}

		resp := s.Commands.Execute(cmd, args, ttl)

		if cmd == "LRANGE" {
			arr := strings.Split(resp, ",")
			s.handleLRANGE(w, arr)
			if err := w.Flush(); err != nil {
				log.Println("flush error:", err)
				return
			}
			continue
		}

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

func (s *Server) handleLRANGE(w *bufio.Writer, arr []string) {
	fmt.Fprintf(w, "*%d\r\n", len(arr)) // send array header

	log.Println(arr)

	for _, val := range arr {
		fmt.Fprintf(w, "$%d\r\n%s\r\n", len(val), val)
	}

}
