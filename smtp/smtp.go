package smtp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Handler func()

type Server struct {
	Addr    string
	Handler Handler
}

func (srv *Server) ListenAndServe() error {
	if srv.Addr == "" {
		srv.Addr = ":25"
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}

	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil {
			// Just a temporary error
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("Accept error: %v", err)
				continue
			}

			return err
		}

		conn := srv.newConn(c)
		go conn.serve()
	}

	return nil
}

// Creat a wrapper around net.Conn
func (srv *Server) newConn(c net.Conn) *Conn {
	return &Conn{
		c: c,
	}
}

type Conn struct {
	c net.Conn
}

func (conn *Conn) serve() error {
	defer conn.c.Close()

	log.Printf("Received new connection")
	fmt.Fprintf(conn.c, "220 SMTP ready\n")

	br := bufio.NewReader(conn.c)
	for {
		line, _ := br.ReadString('\n')

		verb, args := conn.parseLine(line)
		log.Printf("%s -------- %s", verb, args)
	}

	return nil
}

func (conn *Conn) parseLine(line string) (verb string, args string) {
	i := strings.Index(line, " ")
	if i == -1 {
		verb = strings.ToUpper(line)
		return
	}

	verb = strings.ToUpper(line[:i])
	args = line[i+1 : len(line)]
	return
}
