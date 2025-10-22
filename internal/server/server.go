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

		if err != nil {
			fmt.Fprintf(conn, "-ERR%v\r\n", err)
			w.Flush()
			continue
		}

		if cmd == "SUBSCRIBE" {
			if len(args) < 1 {
				fmt.Fprintf(w, "-ERR wrong number of arguments for 'subscribe' command\r\n")
				w.Flush()
				continue
			}

			channel := args[0]
			subChan := s.Commands.GetDB().Subscribe(channel)

			// Correct subscribe confirmation
			fmt.Fprintf(w, "*3\r\n$9\r\nsubscribe\r\n$%d\r\n%s\r\n:1\r\n", len(channel), channel)
			w.Flush()

			log.Printf("client subscribed to channel %s", channel)

			// Start listening for published messages
			go func() {
				for msg := range subChan {
					fmt.Fprintf(w, "*3\r\n$7\r\nmessage\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
						len(channel), channel, len(msg), msg)
					if err := w.Flush(); err != nil {
						log.Println("subscriber flush error:", err)
						s.Commands.GetDB().Unsubscribe(channel, subChan)
						return
					}
				}
			}()
			continue
		}

		if cmd == "UNSUBSCRIBE" {
			if len(args) < 1 {
				fmt.Fprintf(w, "-ERR wrong number of arguments for 'unsubscribe' command\r\n")
				w.Flush()
				continue
			}

			channel := args[0]
			s.Commands.GetDB().Unsubscribe(channel, nil) // you can manage this properly later
			fmt.Fprintf(w, "*3\r\n$11\r\nunsubscribe\r\n$%d\r\n%s\r\n:0\r\n", len(channel), channel)
			w.Flush()
			continue
		}

		if cmd == "PUBLISH" {
			if len(args) < 2 {
				fmt.Fprintf(w, "-ERR wrong number of arguments for 'publish' command\r\n")
				w.Flush()
				continue
			}

			channel, message := args[0], args[1]
			count := s.Commands.GetDB().Publish(channel, message)
			fmt.Fprintf(w, ":%d\r\n", count)
			w.Flush()
			continue
		}

		resp := s.Commands.Execute(cmd, args, ttl)

		if cmd == "LRANGE" || cmd == "SMEMBERS" || cmd == "HGETALL" {
			arr := strings.Split(resp, ",")
			s.handleStringArrays(w, arr)
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

func (s *Server) handleStringArrays(w *bufio.Writer, arr []string) {

	if arr[0] == "+0\r\n" {
		fmt.Fprintf(w, "*0\r\n")
		return
	}

	fmt.Fprintf(w, "*%d\r\n", len(arr)) // send array header

	for _, val := range arr {
		fmt.Fprintf(w, "$%d\r\n%s\r\n", len(val), val)
	}

}
